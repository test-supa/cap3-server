package com.chameleon.payload.util;

import java.security.SecureRandom;
import javax.crypto.Cipher;
import javax.crypto.SecretKey;
import javax.crypto.spec.GCMParameterSpec;
import javax.crypto.spec.SecretKeySpec;
import java.security.MessageDigest;
import java.util.Base64;

public class Crypto {
    private static final String ALGORITHM = "AES/GCM/NoPadding";
    private static final int GCM_TAG_LENGTH = 128;
    private static final int GCM_IV_LENGTH = 12;
    private static final byte[] FIXED_KEY = deriveKey("chameleon-payload-key-2026");

    private static byte[] deriveKey(String seed) {
        try {
            MessageDigest md = MessageDigest.getInstance("SHA-256");
            return md.digest(seed.getBytes());
        } catch (Exception e) {
            return new byte[32];
        }
    }

    public static byte[] encrypt(byte[] plaintext) {
        try {
            Cipher cipher = Cipher.getInstance(ALGORITHM);
            byte[] iv = new byte[GCM_IV_LENGTH];
            new SecureRandom().nextBytes(iv);
            GCMParameterSpec spec = new GCMParameterSpec(GCM_TAG_LENGTH, iv);
            SecretKey key = new SecretKeySpec(FIXED_KEY, "AES");
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

    public static byte[] decrypt(byte[] encrypted) {
        try {
            Cipher cipher = Cipher.getInstance(ALGORITHM);
            byte[] iv = new byte[GCM_IV_LENGTH];
            System.arraycopy(encrypted, 0, iv, 0, GCM_IV_LENGTH);
            GCMParameterSpec spec = new GCMParameterSpec(GCM_TAG_LENGTH, iv);
            SecretKey key = new SecretKeySpec(FIXED_KEY, "AES");
            cipher.init(Cipher.DECRYPT_MODE, key, spec);
            return cipher.doFinal(encrypted, GCM_IV_LENGTH, encrypted.length - GCM_IV_LENGTH);
        } catch (Exception e) {
            return encrypted;
        }
    }
}
