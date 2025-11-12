package org.example;

import android.os.Bundle;
import android.os.Environment;
import android.util.Log;
import androidx.appcompat.app.AppCompatActivity;


import com.OpennhpLibrary;
import com.fancy.zerotrust.R;

import java.io.File;

public class MainActivity extends AppCompatActivity {

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        // Read the phone's storage download directory.
        String appDir = Environment.getExternalStorageDirectory() + File.separator + "download";
        // Does the nhp directory exist in the downloads
        File file = new File(appDir);
        if (!file.exists()) {
            Log.d("MainActivity","download file not exist！");
            return;
        }
        Log.d("MainActivity","download file exist！");
        String appDir1 = Environment.getExternalStorageDirectory() + File.separator + "download"+ File.separator + "nhp";
        boolean initFlag = OpennhpLibrary.INSTANCE.nhp_agent_init(appDir1, 3);
        if (!initFlag) {
            System.out.println("NHP Agent init failed");
            System.exit(0);
        }
        System.out.println("start the loop knocking thread...");
        OpennhpLibrary.INSTANCE.nhp_agent_knockloop_start();
    }
}