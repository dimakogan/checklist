package org.golang.example.bind.test;


import android.content.res.Resources;
import android.util.Log;

import androidx.test.filters.SmallTest;
import androidx.test.platform.app.InstrumentationRegistry;

import org.apache.commons.io.IOUtils;
import org.junit.Assert;
import org.junit.Before;
import org.junit.Test;

import java.io.IOException;
import java.io.InputStream;

import testing.Testing;


@SmallTest
public class BenchmarkTest {

    public class LogPrinter implements testing.Printer {
        private String tag;

        public LogPrinter(String tag) {
            this.tag = tag;
        }
        public void print(String s) {
            Log.d(tag, s);
        }
    }

    private String trace;
    private testing.Printer stdOut, stdErr;

    @Before
    public void setUp() {
        Resources res = InstrumentationRegistry.getInstrumentation().getContext().getResources();
        InputStream is = res.openRawResource(R.raw.trace);
        try {
            trace = IOUtils.toString(is);
        }catch (IOException e) {
            Log.d("Checklist", "benchmark failed:"+e.getMessage());
        }
        stdOut = new LogPrinter("Checklist-StdOut");
        stdErr = new LogPrinter("Checklist-StdErr");
    }

    @Test
    public void runSmokeTest() {
        Log.d("Checklist", "Start smoke test");
    }

    @Test
    public void runBenchmarkPunc() {
        Log.d("Checklist", "Start runBenchmarkPunc");
        Assert.assertTrue(testing.Testing.benchmarkTraceForAndroid(
                trace, Testing.PirTypePunc, stdOut, stdErr));
        Log.d("Checklist", "End runBenchmarkPunc");
    }

   @Test
    public void runBenchmarkDPF() {
        Log.d("Checklist", "Start runBenchmarkDPF");
        Assert.assertTrue(testing.Testing.benchmarkTraceForAndroid(
                trace, Testing.PirTypeDPF, stdOut, stdErr));
        Log.d("Checklist", "End runBenchmarkDPF");
    }

    @Test
    public void runBenchmarkNonPrivate() {
        Log.d("Checklist", "Start runBenchmarkNonPrivate");
        Assert.assertTrue(testing.Testing.benchmarkTraceForAndroid(
                trace, Testing.PirTypeNonPrivate, stdOut, stdErr));
        Log.d("Checklist", "End runBenchmarkNonPrivate");
    }

    @Test
    public void runBaseline() {
        Log.d("Checklist", "Start runBaseline");
        Assert.assertTrue(testing.Testing.benchmarkReference(stdOut, stdErr));
        Log.d("Checklist", "End runBaseline");
    }
}