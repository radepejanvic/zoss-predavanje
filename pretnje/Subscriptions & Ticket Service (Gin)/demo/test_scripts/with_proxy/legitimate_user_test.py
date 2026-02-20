#!/usr/bin/env python3
import requests
import time
import sys

# Server configuration
BASE_URL = "http://localhost:8082"
PROXY_IP = "203.0.113.50"  # IP našeg reverse proxy-ja

def test_legitimate_access():
    print("=" * 70)
    print("👤 LEGITIMATE USER TEST")
    print("=" * 70)
    print(f"\nScenario: Legitimni korisnik pokušava da pristupi aplikaciji")
    print(f"Korisnik dolazi kroz reverse proxy: {PROXY_IP}")
    print(f"Target: {BASE_URL}/api/subscriptions/")
    
    
    # U realnom scenariju, reverse proxy dodaje X-Forwarded-For header
    headers = {
        'X-Forwarded-For': PROXY_IP,  # Automatski dodaje proxy
        'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36'
    }
    
    try:
        response = requests.get(
            BASE_URL + "/api/subscriptions/",
            headers=headers,
            timeout=5
        )
        
        print("=" * 70)
        print("REZULTAT")
        print("=" * 70)
        print(f"Status Code: {response.status_code}")
        print(f"Status Text: {response.reason}")
        
        if response.status_code == 200:
            print("\nSUCCESS: Legitimni korisnik je uspešno pristupio aplikaciji!")
            print("Rate limiter još nije ispunjen ili je već istekao.")
        
        elif response.status_code == 429:
            print("\nBLOCKED: Legitimni korisnik je blokiran!")
            print(f"\nPROBLEM:")
            print(f"   - Napadač je prethodno ispunio rate limiter za IP {PROXY_IP}")
            print(f"   - Rate limiter ne razlikuje napadača od legitimnog korisnika")
            print(f"   - Oba dolaze sa istog IP-a ({PROXY_IP})")
            print(f"   - Rezultat: Legitimni korisnik NE MOŽE da koristi aplikaciju!")
            
            print(f"\nResponse body:")
            try:
                print(f"   {response.json()}")
            except:
                print(f"   {response.text}")
        
        else:
            print(f"\nUnexpected status code: {response.status_code}")
        
    except requests.exceptions.Timeout:
        print("ERROR: Request timeout (server nije odgovorio)")
    except requests.exceptions.ConnectionError:
        print("ERROR: Ne mogu da se povežem na server")
        print(f"   Da li je server pokrenut na {BASE_URL}?")
    except Exception as e:
        print(f"ERROR: {type(e).__name__}: {e}")

if __name__ == "__main__":
    try:
        test_legitimate_access()
    except KeyboardInterrupt:
        print("\n Test stopped by user.")
        sys.exit(0)
