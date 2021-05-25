BROWSER=/Applications/Firefox\ Nightly.app/Contents/MacOS/firefox
# Location of a Firefox profile that needs a safebrowsing update
PROFILE=./ff-profile

TMP=/tmp/`cat /dev/urandom | head -c 32 | shasum | head -c 16`
cp -r $PROFILE $TMP

echo "Using profile in $TMP"
"$BROWSER" --jsconsole --profile $TMP
