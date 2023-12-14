backend server for ninetynine project that handle http request such as authentication and room management

## requirements
have go 1.21.5 installed
have python and requests library install
other dependencies are managed by go modules

add serviceAccountKey.json to root directory

## setup
```bash
go mod tidy
```

## run go server
```bash
.\script.bat
```

## run python client
```bash
python client.py
```