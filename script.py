# for testing purposes
import requests

url = "http://localhost:8080"

# edit this
body_data = {
    "username": "test",
    "password": "123456",
    "email": "test@email.com"
}

# edit this
path = "/login"

if __name__ == '__main__':
    res = requests.post(f"{url}{path}", json=body_data)
    print(res.status_code)
    print(res.json())