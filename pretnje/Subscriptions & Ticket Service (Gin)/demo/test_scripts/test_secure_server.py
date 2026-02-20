#!/usr/bin/env python3
import requests
import time
import sys
from threading import Thread, Lock

# Server configuration
BASE_URL = "http://localhost:8083"
TARGET_ENDPOINT = "/api/subscriptions/"

# Attack configuration
TOTAL_REQUESTS = 50
FAKE_IPS = [f"10.0.0.{i}" for i in range(1, 51)]  # 50 različitih lažnih IP adresa

# Statistics
stats_lock = Lock()
stats = {
    'sent': 0,
    'success': 0,
    'rate_limited': 0,
    'errors': 0,
    'start_time': None
}

def print_stats():
    """Background thread koji prikazuje statistiku"""
    while True:
        time.sleep(2)
        with stats_lock:
            elapsed = time.time() - stats['start_time'] if stats['start_time'] else 0
            print(f"\STATISTICS (after {elapsed:.1f}s):")
            print(f"  Sent: {stats['sent']}")
            print(f"Success (200): {stats['success']}")
            print(f"Rate Limited (429): {stats['rate_limited']}")
            print(f"Errors: {stats['errors']}")

def send_spoofed_request(fake_ip):
    """
    Pokušava da lažira IP adresu sa X-Forwarded-For headerom.
    NA SIGURNOM SERVERU: Server ignorise ovaj header!
    """
    headers = {
        'X-Forwarded-For': fake_ip,
        'User-Agent': 'Attacker-Trying-To-Spoof/1.0'
    }
    
    try:
        response = requests.get(
            BASE_URL + TARGET_ENDPOINT,
            headers=headers,
            timeout=5
        )
        
        with stats_lock:
            stats['sent'] += 1
            
            if response.status_code == 200:
                stats['success'] += 1
            elif response.status_code == 429:
                stats['rate_limited'] += 1
            else:
                stats['errors'] += 1
                
    except requests.exceptions.Timeout:
        with stats_lock:
            stats['sent'] += 1
            stats['errors'] += 1
    except Exception as e:
        with stats_lock:
            stats['sent'] += 1
            stats['errors'] += 1

def main():
    print("=" * 70)
    print("TESTING SECURE SERVER")
    print("=" * 70)
    print(f"\nTarget: {BASE_URL}")
    print(f"Total Requests: {TOTAL_REQUESTS}")
    print(f"Attack Method: Spoofed X-Forwarded-For headers")
    print(f"\nExpected Result: Attack should FAIL")
    print("   Server ignores X-Forwarded-For and uses REAL IP (127.0.0.1)")
    print("   Rate limiter blocks attacker after ~10 requests")
    print("   Other users are NOT affected")
    print("\nStarting test in 3 seconds...")
    print("Press Ctrl+C to stop early.\n")
    
    try:
        time.sleep(3)
    except KeyboardInterrupt:
        sys.exit(0)
    
    # Pokreni background thread za statistiku
    stats['start_time'] = time.time()
    stats_thread = Thread(target=print_stats, daemon=True)
    stats_thread.start()
    
    print("Test started!\n")
    
    try:
        # Pokušaj da lažiraš različite IP adrese
        for i, fake_ip in enumerate(FAKE_IPS[:TOTAL_REQUESTS]):
            send_spoofed_request(fake_ip)
            time.sleep(0.1)
            
    except KeyboardInterrupt:
        print("\nTest interrupted by user (Ctrl+C)")
    
    # Finalna statistika
    elapsed = time.time() - stats['start_time']
    print("\n" + "=" * 70)
    print("FINAL RESULTS")
    print("=" * 70)
    print(f"Duration: {elapsed:.2f} seconds")
    print(f"Total Requests: {stats['sent']}")
    print(f"Successful (200): {stats['success']}")
    print(f"Rate Limited (429): {stats['rate_limited']}")
    print(f"Errors: {stats['errors']}")
    
    print("\n" + "=" * 70)

if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        sys.exit(0)
