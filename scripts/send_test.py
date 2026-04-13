#!/usr/bin/env python3
import hashlib
import hmac
import json
import os
import time
import urllib.request


def main():
    relay_url = os.environ["RELAY_URL"]
    auth_token = os.environ["AUTH_TOKEN"]
    security_level = os.getenv("SECURITY_LEVEL", "basic")
    hmac_secret = os.getenv("HMAC_SECRET", "")

    payload = {
        "title": "Baota test alert",
        "message": "This is a test message from relay integration.",
        "level": "info",
        "source": "baota",
        "event_time": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
        "event_id": f"evt-test-{int(time.time())}",
        "labels": {"env": "test"},
    }
    raw = json.dumps(payload, ensure_ascii=False).encode("utf-8")
    headers = {
        "Content-Type": "application/json",
        "Authorization": f"Bearer {auth_token}",
    }

    if security_level in {"medium", "strict"}:
        if not hmac_secret:
            raise RuntimeError("HMAC_SECRET is required for medium/strict")
        ts = str(int(time.time()))
        sign_input = f"{ts}.{raw.decode('utf-8')}".encode("utf-8")
        signature = hmac.new(hmac_secret.encode("utf-8"), sign_input, hashlib.sha256).hexdigest()
        headers["X-Timestamp"] = ts
        headers["X-Signature"] = signature

    req = urllib.request.Request(relay_url, data=raw, headers=headers, method="POST")
    with urllib.request.urlopen(req, timeout=10) as resp:
        print(resp.status, resp.read().decode("utf-8"))


if __name__ == "__main__":
    main()
