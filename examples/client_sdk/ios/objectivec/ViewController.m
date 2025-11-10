//
//  ViewController.m
//  TestXCFramework
//
//  Created by haochangjiu on 2025/10/30.
//

#import "ViewController.h"
#import <Nhpagent/Nhpagent.h>
#import "FileCopyManager.h"

@interface ViewController ()

@end

@implementation ViewController

- (void)viewDidLoad {
    [super viewDidLoad];
    // Do any additional setup after loading the view.
    // Invoke method to copy files from etc folder to sandbox etc directory
    [FileCopyManager copyFilesToSandboxEtc];
    // Retrieve the sandbox target path (Documents), which is the parent directory of the etc folder
    NSArray *documentsURLs = [[NSFileManager defaultManager] URLsForDirectory:NSDocumentDirectory inDomains:NSUserDomainMask];
    NSURL *documentsURL = [documentsURLs firstObject];
    if (!documentsURL) {
        NSLog(@"Error: Failed to read Documents directory");
    }
    // Get the parent directory path of the etc folder
    NSString *etcPath = documentsURL.path;
    // SdkNhpAgentInit
    BOOL initFlag = IossdkNhpAgentInit(etcPath, 3);
    if (!initFlag) {
        NSLog(@"NHP Agent init failed");
        return;
    }
    // knockloop_start
    long value = IossdkNhpAgentKnockloopStart();
    NSLog(@"SdkNhpAgentKnockloopStart value : %ld", value);
}

@end
