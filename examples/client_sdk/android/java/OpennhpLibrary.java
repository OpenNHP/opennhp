package org.example;

import com.sun.jna.Library;
import com.sun.jna.Native;

/**
 * OpenNHP agent sdk interface
 *
 * @author haochangjiu
 * @version JDK 8
 * @className OpennhpLibrary
 * @date 2025/10/27
 */
public interface OpennhpLibrary extends Library {
    // load OpenNHP agent sdk
    OpennhpLibrary INSTANCE = Native.load("nhpagent", OpennhpLibrary.class);

    /**
     * @description Initialization of the nhp_agent instance working directory path:
     *              The configuration files to be read are located under workingdir/etc/,
     *              and log files will be generated under workingdir/logs/.
     * @param workingDir: the working directory path for the agent
     * @param logLevel:   0: silent, 1: error, 2: info, 3: debug, 4: verbose
     *                    return boolean Whether agent instance has been initialized successfully.
     * @return boolean
     * @author haochangjiu
     * @date 2025/10/27
     * {@link boolean}
     */
    boolean nhp_agent_init(String workingDir, int logLevel);

    /**
     * @description Synchronously stop and release nhp_agent.
     * @author haochangjiu
     * @date 2025/10/27
     */
    void nhp_agent_close();
    /**
     * @description Read the user information, resource information, server information,
     *              and other configuration files written under workingdir/etc,
     *              and asynchronously start the loop knocking thread.
     * @return int
     * @author haochangjiu
     * @date 2025/10/27
     * {@link int}
     */
    int nhp_agent_knockloop_start();

    /**
     * @description Synchronously stop the loop, knock-on sub thread
     * @author hangchangjiu
     * @date 2025/10/27
     */
    void nhp_agent_knockloop_stop();
}