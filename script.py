# import websocket
# import threading

# def on_message(ws, message):
#     print(f"Received: {message}")

# def on_error(ws, error):
#     print(f"Error: {error}")

# def on_close(ws, _, __):
#     print("Connection closed")

# def on_open(ws):
#     print("Connection opened")
#     # Send a message after connection is opened
#     ws.send("Hello, WebSocket server!")

# global ws

# def start_websocket_client():
#     global ws
#     ws = websocket.WebSocketApp("ws://localhost:8080/ws/000000000000",
#                                 on_message=on_message,
#                                 on_error=on_error,
#                                 on_close=on_close)
#     ws.on_open = on_open

#     ws.run_forever()

# if __name__ == "__main__":
#     # Start WebSocket client in a separate thread
#     client_thread = threading.Thread(target=start_websocket_client)
#     client_thread.start()

#     while True:
#         msg = input("Enter message: ")
#         if msg == 'q':
#             ws.close()
#             break
#         else:
#             ws.send(msg)

import requests

url = "http://localhost:8080"

# edit this
body_data = {
    "email": "test@email.com",
    "password": "123456"
}

# edit this
path = "/login"

if __name__ == '__main__':
    res = requests.post(f"{url}{path}", json=body_data)
    print(res.status_code)
    print(res.content)