# frame-light
Simple Webserver which provides an api and ui to switch on lights connected to a raspberry pi

## Installation

Execute the following script to install the application
```bash
DOWNLOAD_URL=$(curl -s https://api.github.com/repos/raynigon/frame-light/releases/latest | \
python3 -c "import sys, json; print(list(filter(lambda x: '$(arch)'[:-1] in x['name'], json.load(sys.stdin)['assets']))[0]['url'])")
wget "--header=Accept: application/octet-stream" -O frame-light.zip $DOWNLOAD_URL
unzip frame-light.zip
```
