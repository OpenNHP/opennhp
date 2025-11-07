//
//  FileCopyManager.m
//  TestXCFramework
//
//  Created by haochangjiu on 2025/10/30.
//

#import "FileCopyManager.h"
#import <Foundation/Foundation.h>

@implementation FileCopyManager

/// Copy the specified file(s) to the etc and certs directories in the application's home directory
+ (void)copyFilesToSandboxEtc {
    // 1. Retrieve the sandboxed Documents directory
    NSArray *documentsURLs = [[NSFileManager defaultManager] URLsForDirectory:NSDocumentDirectory inDomains:NSUserDomainMask];
    NSURL *documentsURL = [documentsURLs firstObject];
    if (!documentsURL) {
        NSLog(@"Failed to retrieve Documents directory");
        return;
    }

    // 2. Define paths for etc and certs directories within the sandbox
    NSURL *etcURL = [documentsURL URLByAppendingPathComponent:@"etc"];
    NSURL *certsURL = [etcURL URLByAppendingPathComponent:@"certs"];

    // 3. Create etc and certs directories (if they don't exist)
    [self createDirectoryIfNotExists:etcURL];
    [self createDirectoryIfNotExists:certsURL];

    // 4. Copy toml files to the etc directory
    NSArray *tomlFiles = @[@"server.toml", @"config.toml", @"dhp.toml", @"resource.toml"];
    for (NSString *fileName in tomlFiles) {
        [self copyFileFromBundle:fileName toDestinationURL:etcURL];
    }

    // 5. Copy certificate files to the etc/certs directory
    NSArray *certFiles = @[@"server.crt", @"server.key"];
    for (NSString *fileName in certFiles) {
        [self copyFileFromBundle:fileName toDestinationURL:certsURL];
    }
}

/// Create directory if it does not exist
+ (void)createDirectoryIfNotExists:(NSURL *)directoryURL {
    NSFileManager *fileManager = [NSFileManager defaultManager];
    if (![fileManager fileExistsAtPath:directoryURL.path]) {
        NSError *error;
        BOOL success = [fileManager createDirectoryAtURL:directoryURL
                              withIntermediateDirectories:YES
                                               attributes:nil
                                                    error:&error];
        if (success) {
            NSLog(@"Directory created successfully: %@", directoryURL.path);
        } else {
            NSLog(@"Failed to create directory: %@, error: %@", directoryURL.path, error.localizedDescription);
        }
    } else {
        NSLog(@"Directory already exists: %@", directoryURL.path);
    }
}

/// Copy file from Bundle to destination path
+ (void)copyFileFromBundle:(NSString *)fileName toDestinationURL:(NSURL *)destinationURL {
    // Get the file path in the Bundle
    NSURL *sourceURL = [[NSBundle mainBundle] URLForResource:[fileName stringByDeletingPathExtension]
                                                withExtension:[fileName pathExtension]];
    if (!sourceURL) {
        NSLog(@"File not found in Bundle: %@", fileName);
        return;
    }

    // Destination file path (destination directory + file name)
    NSURL *destFileURL = [destinationURL URLByAppendingPathComponent:fileName];

    // Copy file (if it doesn't exist)
    NSFileManager *fileManager = [NSFileManager defaultManager];
    if (![fileManager fileExistsAtPath:destFileURL.path]) {
        NSError *error;
        BOOL success = [fileManager copyItemAtURL:sourceURL toURL:destFileURL error:&error];
        if (success) {
            NSLog(@"File copied successfully: %@ -> %@", fileName, destFileURL.path);
        } else {
            NSLog(@"File copy failed: %@, error: %@", fileName, error.localizedDescription);
        }
    } else {
        NSLog(@"File already exists: %@", destFileURL.path);
    }
}

@end
