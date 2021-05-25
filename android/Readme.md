# Runnin Android experiments


## Dependencies
1. Install JDK: either from [Oracle](https://www.oracle.com/java/technologies/javase-jdk16-downloads.html) or using your package manager
2. Create a directory for Android SDK (for example, `make -p ~/Android/sdk`)
3. Set the `ANDROID_HOME` environment variable (for example, `export ANDROID_HOME=~/Android/sdk`)
4. Download [Android SDK tools](https://developer.android.com/studio#span-idcommand-toolsa-namecmdline-toolsacommand-line-tools-onlyspan) and unzip into `$ANDROID_HOME`
5. Install the Android SDK components `build-tools;30.0.3`, `emulator`, `patcher;v4`, `platform-tools`, `platforms;android-29`, `ndk;22.1.7171670`
    ```
    $ cd $ANDROID_HOME
    $ ./cmdline-tools/bin/sdkmanager --sdk_root=. "build-tools;30.0.3" "emulator" "patcher;v4" "platform-tools"  "platforms;android-29" "ndk;22.1.7171670"
    ```

## Build steps
    
1. Generate the Java bindings: 
    ```
    $ cd $CHECKLIST_ROOT
    $ ANDROID_NDK_HOME=$ANDROID_HOME/ndk/22.1.7171670 gomobile bind -o android/app/testing.aar -target android/arm64 ./testing
    ```
2. Build the APK
    ```
    $ # Run from $CHECKLIST_ROOT/android
    $ ANDROID_SDK_ROOT=$ANDROID_HOME ./gradlew build
    ```
3. Build the instrumentation tests
    ```
    $ # Run from $CHECKLIST_ROOT/android
    $ ANDROID_SDK_ROOT=$ANDROID_HOME ./gradlew assembleAndroidTest
    ```
