//
//  FileCopyManager.h
//  TestXCFramework
//
//  Created by haochangjiu on 2025/10/30.
//

#import <Foundation/Foundation.h>

NS_ASSUME_NONNULL_BEGIN

@interface FileCopyManager : NSObject
/// Copy the specified file(s) to the etc and certs directories in the application's home directory
+ (void)copyFilesToSandboxEtc;
@end

NS_ASSUME_NONNULL_END
