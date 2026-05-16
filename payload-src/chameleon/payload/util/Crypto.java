package com.chameleon.payload.util;

import android.util.Base64;
import java.security.MessageDigest;
import java.security.SecureRandom;
import javax.crypto.Cipher;
import javax.crypto.spec.GCMParameterSpec;
import javax.crypto.spec.SecretKeySpec;

public class Crypto {
    private static final String TAG = "Crypto";
    private static final String ALGORITHM = "AES/GCM/NoPadding";
    private static final int GCM_TAG_LENGTH = 128;
    private static final int GCM_IV_LENGTH = 12;

    private static String deviceId = "";
    private static boolean initialized = false;

    public static void init(String devId) {
        deviceId = devId;
        initialized = true;
    }

    private static byte[] deriveKey() {
        try {
            String masterSecret = Secrets.MASTER_SECRET;
            String input = deviceId + ":" + masterSecret;
            MessageDigest digest = MessageDigest.getInstance("SHA-256");
            return digest.digest(input.getBytes("UTF-8"));
        } catch (Exception e) {
            return new byte[32];
        }
    }

    public static byte[] encrypt(byte[] plaintext) {
        if (!initialized) return plaintext;
        try {
            byte[] keyBytes = deriveKey();
            SecretKeySpec key = new SecretKeySpec(keyBytes, "AES");
            Cipher cipher = Cipher.getInstance(ALGORITHM);
            byte[] iv = new byte[GCM_IV_LENGTH];
            new SecureRandom().nextBytes(iv);
            GCMParameterSpec spec = new GCMParameterSpec(GCM_TAG_LENGTH, iv);
            cipher.init(Cipher.ENCRYPT_MODE, key, spec);

            byte[] ciphertext = cipher.doFinal(plaintext);
            byte[] combined = new byte[iv.length + ciphertext.length];
            System.arraycopy(iv, 0, combined, 0, iv.length);
            System.arraycopy(ciphertext, 0, combined, iv.length, ciphertext.length);
            return combined;
        } catch (Exception e) {
            return plaintext;
        }
    }

    public static byte[] decrypt(byte[] ciphertext) {
        if (!initialized) return ciphertext;
        try {
            byte[] keyBytes = deriveKey();
            SecretKeySpec key = new SecretKeySpec(keyBytes, "AES");
            Cipher cipher = Cipher.getInstance(ALGORITHM);

            byte[] iv = new byte[GCM_IV_LENGTH];
            System.arraycopy(ciphertext, 0, iv, 0, GCM_IV_LENGTH);
            GCMParameterSpec spec = new GCMParameterSpec(GCM_TAG_LENGTH, iv);

            cipher.init(Cipher.DECRYPT_MODE, key, spec);
            return cipher.doFinal(ciphertext, GCM_IV_LENGTH, ciphertext.length - GCM_IV_LENGTH);
        } catch (Exception e) {
            return ciphertext;
        }
    }
}
