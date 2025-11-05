package com.example.androidtestsoapp

import android.os.Bundle
import android.os.Environment
import android.util.Log
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.tooling.preview.Preview
import com.example.androidtestsoapp.ui.theme.AndroidTestSoAppTheme
import com.hjq.permissions.Permission
import com.hjq.permissions.XXPermissions
import java.io.File

class MainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        enableEdgeToEdge()
        setContent {
            AndroidTestSoAppTheme {
                Scaffold(modifier = Modifier.fillMaxSize()) { innerPadding ->
                    Greeting(
                        name = "Android",
                        modifier = Modifier.padding(innerPadding)
                    )
                }
            }
        }
        // Request permissions - read/write
        XXPermissions.with(this)
            .permission(Permission.WRITE_EXTERNAL_STORAGE)
            .permission(Permission.READ_MEDIA_IMAGES)
            .permission(Permission.READ_MEDIA_VIDEO)
            .permission(Permission.READ_MEDIA_AUDIO)
            .request { permissions, allGranted ->
                if (allGranted) {
                    Log.d("MainActivity", "Permissions granted")
                    performFileOperations()
                } else {
                    Log.d("MainActivity", "Permissions not granted")
                }
            }
    }
}

/**
 * Need to place the nhp folder containing the etc folder in the phone's download folder
 * After reading the phone storage download directory, call OpennhpLibrary
 */
private fun performFileOperations() {
    // Read phone storage download directory
    val appDir = Environment.getExternalStorageDirectory().toString() + File.separator + "download"
    // Check if zero folder exists in download
    val file = File(appDir)
    if (!file.exists()) {
        Log.d("MainActivity", "Download folder does not exist")
        return
    }
    Log.d("MainActivity", "Download folder exists")
    val appDir1 = Environment.getExternalStorageDirectory().toString() + File.separator + "download" + File.separator + "nhp"
    // Check if nhp folder exists in download
    val file1 = File(appDir1)
    if (!file1.exists()) {
        Log.d("MainActivity", "nhp folder does not exist")
        return
    }
    val appDir2 = Environment.getExternalStorageDirectory().toString() + File.separator + "download" + File.separator + "nhp"+ File.separator + "etc"
    // Check if etc folder exists in download
    val file2 = File(appDir2)
    if (!file2.exists()) {
        Log.d("MainActivity", "Etc folder does not exist")
        return
    }

    val initFlag = OpennhpLibrary.INSTANCE.nhp_agent_init(appDir1, 2)
    if (!initFlag) {
        println("NHP Agent init failed")
        return
    }
    println("start the loop knocking thread...")
    val flag:Int = OpennhpLibrary.INSTANCE.nhp_agent_knockloop_start()
    // Print result
    if (flag > 0) {
        println("NHP Agent knockloop start success")
    } else {
        println("NHP Agent knockloop start failed")
    }
}

@Composable
fun Greeting(name: String, modifier: Modifier = Modifier) {
    Text(
        text = "Hello $name!",
        modifier = modifier
    )
}

@Preview(showBackground = true)
@Composable
fun GreetingPreview() {
    AndroidTestSoAppTheme {
        Greeting("Android")
    }
}