package com.chameleon.payload.overlay;

import android.content.Context;
import android.content.Intent;
import android.util.Log;

public class OverlayManager {
    private static final String TAG = "OverlayManager";
    private final Context context;

    public OverlayManager(Context context) {
        this.context = context;
    }

    public void showPhishingOverlay(String targetApp) {
        Log.i(TAG, "Showing phishing overlay for: " + targetApp);
        // In production: launch overlay activity matching the target app's login screen
        // For now, this is a placeholder that logs the intent
    }

    public void hideOverlay() {
        Log.i(TAG, "Hiding overlay");
    }
}
