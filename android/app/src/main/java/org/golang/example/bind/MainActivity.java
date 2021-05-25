/*
 * Copyright 2015 The Go Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package org.golang.example.bind;

import android.app.Activity;
import android.content.res.Resources;
import android.os.Bundle;
import android.text.method.ScrollingMovementMethod;
import android.widget.TextView;
import org.apache.commons.io.IOUtils;

import org.w3c.dom.Text;

import java.io.IOException;
import java.io.InputStream;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;

import testing.Testing;

public class MainActivity extends Activity {

    public class TextViewPrinter implements testing.Printer {
        private TextView view;
        TextViewPrinter(TextView view) {
            this.view = view;
        }

        public void print(String s) {
            MainActivity.super.runOnUiThread(new Runnable() {
                @Override
                public void run() {view.setText(s);}});
        }
    }

    public class TextViewAppendPrinter implements testing.Printer {
        private TextView view;
        TextViewAppendPrinter(TextView view) {
            this.view = view;
        }

        public void print(String s) {
            MainActivity.super.runOnUiThread(new Runnable() {
                @Override
                public void run() {view.append(s);}});
        }
    }


    private TextView outTextView,errTextView;
    private ExecutorService executor;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.activity_main);
        errTextView = (TextView) findViewById(R.id.errView);
        errTextView.setText("Start");
        testing.Printer errPrinter = new TextViewPrinter(errTextView);
        outTextView = (TextView) findViewById(R.id.outView);
        outTextView.setText("Stdout");
        outTextView.setMovementMethod(new ScrollingMovementMethod());
        testing.Printer outPrinter = new TextViewAppendPrinter(outTextView);

        executor = Executors.newSingleThreadExecutor();

        Resources res = getResources();
        InputStream is = res.openRawResource(R.raw.trace);
        try {
            String trace = IOUtils.toString(is);
            // Call Go function.
            executor.execute(new Runnable() {
                @Override
                public void run() {
                    testing.Testing.benchmarkTraceForAndroid(trace, Testing.PirTypePunc, outPrinter, errPrinter);
                }
            });
            } catch (IOException e) {
            errTextView.setText("Failed to load resource");
        }
        IOUtils.closeQuietly(is);
    }
}
