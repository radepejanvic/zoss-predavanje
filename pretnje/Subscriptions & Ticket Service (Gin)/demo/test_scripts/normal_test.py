#!/usr/bin/env python3

import requests
import time
import sys
from datetime import datetime


TARGET_URL = "http://localhost:8081/api/subscriptions/"
NUM_REQUESTS = 20  # Šaljemo 20 zahteva (rate limit je 10)


def send_normal_request(request_num):
    payload = {
        "user_email": f"normal_user{request_num}@example.com",
        "plan_type": "premium"
    }
    
    try:
        response = requests.post(TARGET_URL, json=payload, timeout=5)
        timestamp = datetime.now().strftime("%H:%M:%S")
        
        if response.status_code == 201:
            print(f"{timestamp}] Request #{request_num} - SUCCESS (Status: {response.status_code})")
            return True
        elif response.status_code == 429:
            print(f"[{timestamp}] Request #{request_num} - BLOCKED - Rate limit exceeded!")
            if response.headers.get('Content-Type') == 'application/json':
                print(f"    Response: {response.json()}")
            return False
        else:
            print(f"[{timestamp}] Request #{request_num} - ERROR (Status: {response.status_code})")
            return False
            
    except requests.exceptions.RequestException as e:
        print(f"Request #{request_num} - Connection Error: {str(e)}")
        return False


def run_normal_test():
    print("=" * 70)
    print("Normal Rate Limiter Test")
    print("=" * 70)
    print(f"Target: {TARGET_URL}")
    print(f"Total requests: {NUM_REQUESTS}")
    print(f"Expected limit: 10 requests per minute")
    print("=" * 70)
    print()
    
    try:
        response = requests.get("http://localhost:8080/api/subscriptions/", timeout=2)
        if response.status_code == 200:
            print("Server is reachable")
        else:
            print("Server responded with unexpected status")
    except:
        print("ERROR: Cannot reach server! Make sure the Gin app is running.")
        return
    
    print()
    print("Starting test...")
    print()
    
    successful = 0
    blocked = 0
    
    for i in range(1, NUM_REQUESTS + 1):
        success = send_normal_request(i)
        if success:
            successful += 1
        else:
            blocked += 1
        
        time.sleep(0.5)  # kratka pauza između zahteva
    
    print()
    print("=" * 70)
    print("Test Results:")
    print(f"Successful: {successful}")
    print(f"Blocked: {blocked}")
    print("=" * 70)
    print()
    
    if blocked > 0:
        print("Rate limiter is working correctly!")
        print("Requests were blocked after exceeding the limit.")
    else:
        print("Warning: No requests were blocked - rate limiter may not be active.")
    print()


if __name__ == "__main__":
    try:
        run_normal_test()
    except KeyboardInterrupt:
        print("\nTest stopped by user (Ctrl+C)!")
        sys.exit(0)
