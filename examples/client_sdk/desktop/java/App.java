package org.example;

import java.util.Scanner;

/**
 * Application for calling the OpenNHP agent SDK
 *
 * @author haochangjiu
 * @version JDK 8
 * @className App
 * @date 2025/10/27
 */
public class App {
    public static void main(String[] args) throws Exception {
//        Initialize and start the OpenNHP agent SDK service
        boolean initFlag = OpennhpLibrary.INSTANCE.nhp_agent_init("D:\\console-workspace\\opennhp-knock", 3);
        if (!initFlag) {
            System.out.println("NHP Agent init failed");
            System.exit(0);
        }
//        Invoke methods in the OpenNHP agent SDK via input commands
        Scanner scanner = new Scanner(System.in);

        while (true) {
            System.out.print("> ");
            if (scanner.hasNextLine()) {
                String input = scanner.nextLine().trim();
                if ("knock".equalsIgnoreCase(input)) {
                    System.out.println("start the loop knocking thread...");
                    OpennhpLibrary.INSTANCE.nhp_agent_knockloop_start();
                } else if ("cancel".equalsIgnoreCase(input)) {
                    System.out.println("stop the loop knocking thread...");
                    OpennhpLibrary.INSTANCE.nhp_agent_knockloop_stop();
                } else if ("exit".equalsIgnoreCase(input)) {
                    System.out.println("exit nhp agent service...");
                    OpennhpLibrary.INSTANCE.nhp_agent_close();
                    break;
                } else {
                    System.out.println("invalid input");
                }
            }
        }
        scanner.close();
    }
}