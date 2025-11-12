//
//  ViewController.swift
//  TestXCFrameworkSwift
//
//  Created by haochangjiu on 2025/10/30.
//

import UIKit
import Nhpagent

class ViewController: UIViewController {
    override func viewDidLoad() {
        super.viewDidLoad()
        // Do any additional setup after loading the view.
        // Call method to copy files from etc folder to sandbox etc directory
        FileCopyManager.copyFilesToSandboxEtc()
        // Retrieve the sandbox target path (Documents), which is the parent directory of the etc folder
        guard let documentsURL = FileManager.default.urls(for: .documentDirectory, in: .userDomainMask).first else {
            print("Error: Failed to read Documents directory")
            return
        }
        // Get the parent directory path of the etc folder
        let etcPath: String = documentsURL.path
        // Call SdkNhpAgentInit for initialization
        let initFlag: Bool = IossdkNhpAgentInit(etcPath, 3)
        if !initFlag {
            print("NHP Agent init failed")
        }
        // Call knockloop_start
        let value = IossdkNhpAgentKnockloopStart()
        print("SdkNhpAgentKnockloopStart value: %ld", value)
  }
}