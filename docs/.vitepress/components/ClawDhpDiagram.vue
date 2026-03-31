<template>
  <div class="claw-dhp-wrapper">
    <!-- Legend -->
    <div class="arch-toolbar">
      <div class="legend-bar">
        <span class="legend-item"><span class="legend-line nhp"></span> NHP协议</span>
        <span class="legend-item"><span class="legend-line data"></span> 数据流</span>
        <span class="legend-item"><span class="legend-line task-blue"></span> 任务流</span>
        <span class="legend-item"><span class="legend-line encrypt"></span> 加密/入库</span>
        <span class="legend-item"><span class="legend-line data-red"></span> 数据获取</span>
        <span class="legend-sep">|</span>
        <span class="legend-item"><span class="legend-dot req"></span> 请求</span>
        <span class="legend-item"><span class="legend-ring"></span> 响应</span>
      </div>
    </div>

    <svg viewBox="-20 0 1080 820" xmlns="http://www.w3.org/2000/svg" class="arch-svg"
      role="img" aria-label="OpenClaw + DHP 安全增强架构图：展示AI任务在可信执行环境中通过DHP安全访问加密数据的完整流程">
      <defs>
        <!-- Gradients -->
        <linearGradient id="cdh-grad-node" x1="0" y1="0" x2="1" y2="1">
          <stop offset="0%" stop-color="var(--node-fill-from)" /><stop offset="100%" stop-color="var(--node-fill-to)" />
        </linearGradient>

        <!-- Filters -->
        <filter id="cdh-shadow-sm" x="-4%" y="-4%" width="108%" height="112%">
          <feDropShadow dx="0" dy="1" stdDeviation="2" flood-opacity="var(--shadow-opacity, 0.08)" />
        </filter>
        <filter id="cdh-shadow-md" x="-6%" y="-6%" width="112%" height="120%">
          <feDropShadow dx="0" dy="2" stdDeviation="4" flood-opacity="var(--shadow-opacity-md, 0.12)" />
        </filter>
        <filter id="cdh-shadow-badge" x="-30%" y="-30%" width="160%" height="160%">
          <feDropShadow dx="0" dy="1" stdDeviation="3" flood-opacity="0.25" />
        </filter>
        <!-- Particle glow filters -->
        <filter id="cdh-glow-green" x="-100%" y="-100%" width="300%" height="300%">
          <feGaussianBlur stdDeviation="3" result="blur" />
          <feFlood flood-color="#16a34a" flood-opacity="0.5" result="color" />
          <feComposite in="color" in2="blur" operator="in" result="glow" />
          <feMerge><feMergeNode in="glow" /><feMergeNode in="SourceGraphic" /></feMerge>
        </filter>
        <filter id="cdh-glow-teal" x="-100%" y="-100%" width="300%" height="300%">
          <feGaussianBlur stdDeviation="3" result="blur" />
          <feFlood flood-color="#0d9488" flood-opacity="0.5" result="color" />
          <feComposite in="color" in2="blur" operator="in" result="glow" />
          <feMerge><feMergeNode in="glow" /><feMergeNode in="SourceGraphic" /></feMerge>
        </filter>
        <filter id="cdh-glow-blue" x="-100%" y="-100%" width="300%" height="300%">
          <feGaussianBlur stdDeviation="3" result="blur" />
          <feFlood flood-color="#3b82f6" flood-opacity="0.5" result="color" />
          <feComposite in="color" in2="blur" operator="in" result="glow" />
          <feMerge><feMergeNode in="glow" /><feMergeNode in="SourceGraphic" /></feMerge>
        </filter>
        <filter id="cdh-glow-dark" x="-100%" y="-100%" width="300%" height="300%">
          <feGaussianBlur stdDeviation="3" result="blur" />
          <feFlood flood-color="#475569" flood-opacity="0.4" result="color" />
          <feComposite in="color" in2="blur" operator="in" result="glow" />
          <feMerge><feMergeNode in="glow" /><feMergeNode in="SourceGraphic" /></feMerge>
        </filter>

        <!-- Arrow markers -->
        <marker id="cdh-arr-blue" viewBox="0 0 8 6" refX="7" refY="3" markerWidth="8" markerHeight="6" orient="auto-start-reverse">
          <path d="M0,0.5 L7,3 L0,5.5" fill="none" stroke="var(--arrow-blue)" stroke-width="1.2" stroke-linejoin="round" /></marker>
        <marker id="cdh-arr-green" viewBox="0 0 8 6" refX="7" refY="3" markerWidth="8" markerHeight="6" orient="auto-start-reverse">
          <path d="M0,0.5 L7,3 L0,5.5" fill="none" stroke="var(--arrow-green)" stroke-width="1.2" stroke-linejoin="round" /></marker>
        <marker id="cdh-arr-teal" viewBox="0 0 8 6" refX="7" refY="3" markerWidth="8" markerHeight="6" orient="auto-start-reverse">
          <path d="M0,0.5 L7,3 L0,5.5" fill="none" stroke="var(--arrow-teal)" stroke-width="1.2" stroke-linejoin="round" /></marker>
        <marker id="cdh-arr-dark" viewBox="0 0 8 6" refX="7" refY="3" markerWidth="8" markerHeight="6" orient="auto-start-reverse">
          <path d="M0,0.5 L7,3 L0,5.5" fill="none" stroke="var(--arrow-dark)" stroke-width="1.2" stroke-linejoin="round" /></marker>
        <marker id="cdh-arr-red" viewBox="0 0 8 6" refX="7" refY="3" markerWidth="8" markerHeight="6" orient="auto-start-reverse">
          <path d="M0,0.5 L7,3 L0,5.5" fill="none" stroke="var(--arrow-red)" stroke-width="1.2" stroke-linejoin="round" /></marker>
        <marker id="cdh-arr-store" viewBox="0 0 8 6" refX="7" refY="3" markerWidth="8" markerHeight="6" orient="auto-start-reverse">
          <path d="M0,0.5 L7,3 L0,5.5" fill="none" stroke="var(--arrow-store)" stroke-width="1.2" stroke-linejoin="round" /></marker>

        <!-- Animated connection paths -->
        <!-- (1) User → OpenClaw -->
        <path id="cdh-path-1" d="M 97,478 L 97,598" />
        <!-- (2) OpenClaw → AI Engine -->
        <path id="cdh-path-2" d="M 97,656 L 97,742 L 260,742" />
        <!-- 动态密钥 → NHP代理 (提供密钥, not a numbered step) -->
        <path id="cdh-path-dynkey-agent" d="M 112,312 L 112,395 L 270,395 L 270,412" />
        <!-- (5) NHP服务器 → NHP代理 (请求授权, solid) — left line, straight down -->
        <path id="cdh-path-5" d="M 305,312 L 305,412" />
        <!-- (3) NHP服务器 → NHP代理 (请求远程证明, solid) — right line, straight down -->
        <path id="cdh-path-3" d="M 350,312 L 350,412" />
        <!-- (4) 远程证明服务 → NHP代理 (返回证明结果, solid) -->
        <path id="cdh-path-4" d="M 430,312 L 430,360 L 400,360 L 400,412" />
        <!-- (6) 远程证明服务 → NHP代理 (返回授权, dashed) -->
        <path id="cdh-path-6" d="M 495,312 L 495,375 L 450,375 L 450,412" />
        <!-- Right-side data flow paths (加密 + 入库) -->
        <path id="cdh-path-encrypt-doc"   d="M 680,102 L 680,235" />
        <path id="cdh-path-encrypt-email" d="M 810,102 L 810,235" />
        <path id="cdh-path-encrypt-photo" d="M 940,102 L 940,235" />
        <path id="cdh-path-store-doc"   d="M 680,285 L 680,318 L 750,350" />
        <path id="cdh-path-store-email" d="M 810,285 L 810,350" />
        <path id="cdh-path-store-photo" d="M 940,285 L 940,318 L 870,350" />
        <path id="cdh-path-dhp-ztdo" d="M 820,415 L 820,470" />
        <!-- (7) NHP代理 → ZTDO (获取加密ZTDO) -->
        <path id="cdh-path-7" d="M 480,440 L 578,440 L 578,510 L 700,510" />
        <!-- (8) NHP代理 → 解密后敏感数据 (TEE内解密) -->
        <path id="cdh-path-8" d="M 370,465 L 370,530" />
        <!-- (9) 解密后敏感数据 → AI任务引擎 (供AI使用) -->
        <path id="cdh-path-9" d="M 370,595 L 370,710" />
      </defs>

      <!-- ===== TITLE ===== -->
      <text x="290" y="100" class="diagram-title" text-anchor="middle">OpenClaw + DHP 安全增强架构图</text>

      <!-- ===== ZONE: 用户敏感数据源 (top-right, red border) ===== -->
      <g class="zone zone-datasource">
        <rect x="590" y="18" width="460" height="148" rx="8" class="zone-border-red" />
        <text x="820" y="38" class="zone-title-inner zone-title-red">用户敏感数据源</text>
        <!-- 文档 -->
        <g class="node" :class="{ active: active === 'datasource' }" @mouseenter="active = 'datasource'" @mouseleave="active = null">
          <rect x="625" y="52" width="110" height="50" rx="6" class="box-plain" filter="url(#cdh-shadow-sm)" />
          <text x="680" y="82" class="box-text">文档</text>
        </g>
        <!-- 邮件 -->
        <rect x="755" y="52" width="110" height="50" rx="6" class="box-plain" filter="url(#cdh-shadow-sm)" />
        <text x="810" y="82" class="box-text">邮件</text>
        <!-- 照片 -->
        <rect x="885" y="52" width="110" height="50" rx="6" class="box-plain" filter="url(#cdh-shadow-sm)" />
        <text x="940" y="82" class="box-text">照片</text>
        <!-- 加密 arrows: 文档→加密文档, 邮件→加密邮件, 照片→加密照片 -->
        <text x="660" y="138" class="encrypt-label">加密</text>
        <line x1="680" y1="102" x2="680" y2="165" class="arrow-encrypt" />
        <line x1="680" y1="165" x2="680" y2="235" class="arrow-encrypt" marker-end="url(#cdh-arr-red)" />
        <text x="790" y="138" class="encrypt-label">加密</text>
        <line x1="810" y1="102" x2="810" y2="165" class="arrow-encrypt" />
        <line x1="810" y1="165" x2="810" y2="235" class="arrow-encrypt" marker-end="url(#cdh-arr-red)" />
        <text x="920" y="138" class="encrypt-label">加密</text>
        <line x1="940" y1="102" x2="940" y2="165" class="arrow-encrypt" />
        <line x1="940" y1="165" x2="940" y2="235" class="arrow-encrypt" marker-end="url(#cdh-arr-red)" />
      </g>

      <!-- ===== ZONE: NHP服务器/集群 (blue border) ===== -->
      <g class="zone zone-nhp-cluster">
        <rect x="20" y="200" width="540" height="130" rx="8" class="zone-border-blue" />
        <text x="290" y="218" class="zone-title-inner zone-title-blue">NHP服务器/集群</text>
        <!-- 动态密钥访问对象 -->
        <g class="node" :class="{ active: active === 'dynkey' }" @mouseenter="active = 'dynkey'" @mouseleave="active = null">
          <rect x="35" y="237" width="155" height="75" rx="8" class="box-white" filter="url(#cdh-shadow-sm)" />
          <text x="112" y="268" class="box-text-sm">动态密钥访问对象</text>
          <text x="112" y="286" class="box-text-sub">（一用一密）</text>
        </g>
        <!-- NHP服务器 -->
        <g class="node" :class="{ active: active === 'nhpserver' }" @mouseenter="active = 'nhpserver'" @mouseleave="active = null">
          <rect x="210" y="237" width="155" height="75" rx="8" class="box-white" filter="url(#cdh-shadow-sm)" />
          <text x="287" y="268" class="box-text-sm">NHP服务器</text>
          <text x="287" y="286" class="box-text-sub">(NHP Server)</text>
        </g>
        <!-- 远程证明服务 -->
        <g class="node" :class="{ active: active === 'attestation' }" @mouseenter="active = 'attestation'" @mouseleave="active = null">
          <rect x="385" y="237" width="155" height="75" rx="8" class="box-white" filter="url(#cdh-shadow-sm)" />
          <text x="462" y="275" class="box-text-sm">远程证明服务</text>
        </g>
        <!-- 提供密钥 label is on the dynkey→agent path instead -->
      </g>

      <!-- ===== ZONE: DHP数据经纪件 NHP-DB (red border, tall) ===== -->
      <g class="zone zone-dhp">
        <rect x="590" y="200" width="460" height="440" rx="8" class="zone-border-red" />
        <text x="820" y="220" class="zone-title-inner zone-title-red">DHP数据经纪件（NHP-DB）</text>
        <!-- 加密文档 -->
        <rect x="625" y="237" width="110" height="48" rx="6" class="box-pink" filter="url(#cdh-shadow-sm)" />
        <text x="680" y="266" class="box-text-sm">加密文档</text>
        <!-- 加密邮件 -->
        <rect x="755" y="237" width="110" height="48" rx="6" class="box-pink" filter="url(#cdh-shadow-sm)" />
        <text x="810" y="266" class="box-text-sm">加密邮件</text>
        <!-- 加密照片 -->
        <rect x="885" y="237" width="110" height="48" rx="6" class="box-pink" filter="url(#cdh-shadow-sm)" />
        <text x="940" y="266" class="box-text-sm">加密照片</text>
        <!-- 入库 arrows: 加密文档/邮件/照片 → DHP数据经纪件 (red) -->
        <text x="660" y="308" class="store-label-red">入库</text>
        <path d="M 680,285 L 680,318 L 750,350" class="arrow-red-dashed" fill="none" marker-end="url(#cdh-arr-red)" />
        <text x="790" y="312" class="store-label-red">入库</text>
        <line x1="810" y1="285" x2="810" y2="350" class="arrow-red-dashed" marker-end="url(#cdh-arr-red)" />
        <text x="920" y="308" class="store-label-red">入库</text>
        <path d="M 940,285 L 940,318 L 870,350" class="arrow-red-dashed" fill="none" marker-end="url(#cdh-arr-red)" />

        <!-- DHP数据经纪件 center area -->
        <g class="node" :class="{ active: active === 'nhpdb' }" @mouseenter="active = 'nhpdb'" @mouseleave="active = null">
          <rect x="680" y="350" width="280" height="65" rx="8" class="box-white" filter="url(#cdh-shadow-sm)" />
          <text x="820" y="390" class="box-text">DHP数据经纪件</text>
        </g>

        <!-- Arrow from DHP to ZTDO with 封装ZTDO label on the line -->
        <line x1="820" y1="415" x2="820" y2="470" class="arrow-store" marker-end="url(#cdh-arr-store)" />
        <rect x="782" y="432" width="76" height="18" rx="4" class="ztdo-label-bg" />
        <text x="820" y="445" class="store-label">封装ZTDO</text>

        <!-- ZTDO box -->
        <g class="node" :class="{ active: active === 'ztdo' }" @mouseenter="active = 'ztdo'" @mouseleave="active = null">
          <rect x="700" y="472" width="240" height="75" rx="10" class="box-ztdo" filter="url(#cdh-shadow-sm)" />
          <text x="820" y="507" class="ztdo-title">零信任数据对象</text>
          <text x="820" y="527" class="ztdo-subtitle">ZTDO</text>
        </g>
      </g>

      <!-- ===== ZONE: 可信执行环境 TEE (orange thick border, includes AI engine) ===== -->
      <g class="zone zone-tee">
        <rect x="185" y="380" width="380" height="420" rx="10" class="tee-border" />
        <text x="375" y="400" class="tee-title">可信执行环境 TEE</text>

        <!-- NHP代理 (NHP Agent) -->
        <g class="node" :class="{ active: active === 'agent' }" @mouseenter="active = 'agent'" @mouseleave="active = null">
          <rect x="260" y="412" width="220" height="53" rx="8" class="box-agent" filter="url(#cdh-shadow-sm)" />
          <text x="370" y="436" class="box-text-white">NHP代理</text>
          <text x="370" y="453" class="box-text-white-sub">(NHP Agent)</text>
        </g>

        <!-- 解密后敏感数据（内存中） -->
        <g class="node" :class="{ active: active === 'decrypted' }" @mouseenter="active = 'decrypted'" @mouseleave="active = null">
          <rect x="260" y="530" width="220" height="65" rx="8" class="box-decrypted" filter="url(#cdh-shadow-sm)" />
          <text x="370" y="558" class="box-text-green">解密后敏感数据</text>
          <text x="370" y="576" class="box-text-green-sub">（内存中）</text>
        </g>

        <!-- AI 任务引擎 (inside TEE, single frame) -->
        <g class="node" :class="{ active: active === 'aitask' }" @mouseenter="active = 'aitask'" @mouseleave="active = null">
          <rect x="260" y="710" width="220" height="65" rx="8" class="box-ai-engine" filter="url(#cdh-shadow-sm)" />
          <text x="370" y="738" class="box-text-ai">AI 任务引擎</text>
          <text x="370" y="756" class="box-text-ai-sub">OpenClaw Worker</text>
        </g>
      </g>

      <!-- ===== 用户（发起AI任务）===== -->
      <g class="node" :class="{ active: active === 'user' }" @mouseenter="active = 'user'" @mouseleave="active = null">
        <rect x="30" y="420" width="135" height="58" rx="8" class="box-blue-border" filter="url(#cdh-shadow-sm)" />
        <text x="97" y="446" class="box-text-sm">用户</text>
        <text x="97" y="464" class="box-text-sub">（发起AI任务）</text>
      </g>

      <!-- ===== OpenClaw AI任务调度 ===== -->
      <g class="node" :class="{ active: active === 'openclaw' }" @mouseenter="active = 'openclaw'" @mouseleave="active = null">
        <rect x="30" y="598" width="135" height="58" rx="8" class="box-blue-border" filter="url(#cdh-shadow-sm)" />
        <text x="97" y="622" class="box-text-sm">OpenClaw</text>
        <text x="97" y="640" class="box-text-sub">AI任务调度</text>
      </g>

      <!-- AI 任务引擎 is now inside the TEE zone above -->

      <!-- ===== CONNECTIONS (static arrows) ===== -->
      <g class="connections">
        <!-- (1) User → OpenClaw (blue, task flow) -->
        <use href="#cdh-path-1" class="conn conn-blue" marker-end="url(#cdh-arr-blue)" />
        <!-- (2) OpenClaw → AI Engine (blue, task flow) -->
        <use href="#cdh-path-2" class="conn conn-blue" marker-end="url(#cdh-arr-blue)" />
        <!-- (3) NHP服务器 → NHP代理 (请求远程证明, solid) -->
        <use href="#cdh-path-3" class="conn conn-nhp" marker-end="url(#cdh-arr-green)" />
        <!-- (5) NHP服务器 → NHP代理 (请求授权, solid) -->
        <use href="#cdh-path-5" class="conn conn-nhp" marker-end="url(#cdh-arr-green)" />
        <!-- (4) 远程证明服务 → NHP代理 (返回证明结果, downward, solid) -->
        <use href="#cdh-path-4" class="conn conn-nhp" marker-end="url(#cdh-arr-green)" />
        <!-- (6) NHP服务器 → NHP代理 (返回授权, downward, dashed) -->
        <use href="#cdh-path-6" class="conn conn-nhp conn-nhp-dashed" marker-end="url(#cdh-arr-green)" />
        <!-- 动态密钥 → NHP代理 (提供密钥, blue, NOT a numbered step) -->
        <use href="#cdh-path-dynkey-agent" class="conn conn-blue" marker-end="url(#cdh-arr-blue)" />
        <rect x="155" y="385" width="62" height="16" rx="3" class="step-label-bg" />
        <text x="186" y="397" class="label-blue">提供密钥</text>
        <!-- (7) NHP代理 → ZTDO (red dashed) -->
        <use href="#cdh-path-7" class="conn conn-red-dashed" marker-end="url(#cdh-arr-red)" />
        <!-- (8) Agent → Decrypted (teal, TEE decrypt) -->
        <use href="#cdh-path-8" class="conn conn-data" marker-end="url(#cdh-arr-teal)" />
        <!-- (9) Decrypted → AI Engine (teal, supply AI) -->
        <use href="#cdh-path-9" class="conn conn-data" marker-end="url(#cdh-arr-teal)" />
      </g>

      <!-- ===== PARTICLES (cascading timing matches step order) ===== -->
      <g class="particles-layer">
        <!-- (1) User → OpenClaw (blue) -->
        <circle r="4" class="particle particle-blue" filter="url(#cdh-glow-blue)">
          <animateMotion dur="1.8s" repeatCount="indefinite" keyPoints="0;1" keyTimes="0;1" calcMode="linear"><mpath href="#cdh-path-1" /></animateMotion>
        </circle>

        <!-- (2) OpenClaw → AI Engine (blue) -->
        <circle r="4" class="particle particle-blue" filter="url(#cdh-glow-blue)">
          <animateMotion dur="2.6s"repeatCount="indefinite" keyPoints="0;1" keyTimes="0;1" calcMode="linear"><mpath href="#cdh-path-2" /></animateMotion>
        </circle>

        <!-- (3) NHP服务器 → NHP代理 (请求远程证明) -->
        <circle r="4.5" class="particle particle-green" filter="url(#cdh-glow-green)">
          <animateMotion dur="2.4s"repeatCount="indefinite" keyPoints="0;1" keyTimes="0;1" calcMode="linear"><mpath href="#cdh-path-3" /></animateMotion>
        </circle>

        <!-- (5) NHP服务器 → NHP代理 (请求授权) -->
        <circle r="4.5" class="particle particle-green" filter="url(#cdh-glow-green)">
          <animateMotion dur="2.4s"repeatCount="indefinite" keyPoints="0;1" keyTimes="0;1" calcMode="linear"><mpath href="#cdh-path-5" /></animateMotion>
        </circle>

        <!-- (4) 远程证明服务 → NHP代理 (返回证明结果) -->
        <circle r="4" class="particle particle-green-ring" filter="url(#cdh-glow-green)">
          <animateMotion dur="2.4s"repeatCount="indefinite" keyPoints="0;1" keyTimes="0;1" calcMode="linear"><mpath href="#cdh-path-4" /></animateMotion>
        </circle>

        <!-- (6) 远程证明服务 → NHP代理 (返回授权, dashed) -->
        <circle r="4" class="particle particle-green-ring" filter="url(#cdh-glow-green)">
          <animateMotion dur="2.6s"repeatCount="indefinite" keyPoints="0;1" keyTimes="0;1" calcMode="linear"><mpath href="#cdh-path-6" /></animateMotion>
        </circle>

        <!-- 动态密钥 → NHP代理 (提供密钥, blue) -->
        <circle r="4" class="particle particle-blue" filter="url(#cdh-glow-blue)">
          <animateMotion dur="3.2s"repeatCount="indefinite" keyPoints="0;1" keyTimes="0;1" calcMode="linear"><mpath href="#cdh-path-dynkey-agent" /></animateMotion>
        </circle>

        <!-- (7) NHP代理 → ZTDO (获取加密ZTDO, red) -->
        <circle r="5" class="particle particle-red" filter="url(#cdh-glow-dark)">
          <animateMotion dur="3s"repeatCount="indefinite" keyPoints="0;1" keyTimes="0;1" calcMode="linear"><mpath href="#cdh-path-7" /></animateMotion>
        </circle>

        <!-- (8) NHP代理 → 解密后敏感数据 (TEE内解密) -->
        <circle r="4.5" class="particle particle-teal" filter="url(#cdh-glow-teal)">
          <animateMotion dur="1.4s"repeatCount="indefinite" keyPoints="0;1" keyTimes="0;1" calcMode="linear"><mpath href="#cdh-path-8" /></animateMotion>
        </circle>

        <!-- (9) 解密后敏感数据 → AI任务引擎 (供AI使用) -->
        <circle r="4" class="particle particle-teal" filter="url(#cdh-glow-teal)">
          <animateMotion dur="2s"repeatCount="indefinite" keyPoints="0;1" keyTimes="0;1" calcMode="linear"><mpath href="#cdh-path-9" /></animateMotion>
        </circle>

        <!-- Right-side: 加密 particles (red) -->
        <circle r="3.5" class="particle particle-red" filter="url(#cdh-glow-dark)">
          <animateMotion dur="2s" repeatCount="indefinite" keyPoints="0;1" keyTimes="0;1" calcMode="linear"><mpath href="#cdh-path-encrypt-doc" /></animateMotion>
        </circle>
        <circle r="3.5" class="particle particle-red" filter="url(#cdh-glow-dark)">
          <animateMotion dur="2s"repeatCount="indefinite" keyPoints="0;1" keyTimes="0;1" calcMode="linear"><mpath href="#cdh-path-encrypt-email" /></animateMotion>
        </circle>
        <circle r="3.5" class="particle particle-red" filter="url(#cdh-glow-dark)">
          <animateMotion dur="2s"repeatCount="indefinite" keyPoints="0;1" keyTimes="0;1" calcMode="linear"><mpath href="#cdh-path-encrypt-photo" /></animateMotion>
        </circle>

        <!-- Right-side: 入库 particles (red) -->
        <circle r="3.5" class="particle particle-red" filter="url(#cdh-glow-dark)">
          <animateMotion dur="1.6s" repeatCount="indefinite" keyPoints="0;1" keyTimes="0;1" calcMode="linear"><mpath href="#cdh-path-store-doc" /></animateMotion>
        </circle>
        <circle r="3.5" class="particle particle-red" filter="url(#cdh-glow-dark)">
          <animateMotion dur="1.6s" repeatCount="indefinite" keyPoints="0;1" keyTimes="0;1" calcMode="linear"><mpath href="#cdh-path-store-email" /></animateMotion>
        </circle>
        <circle r="3.5" class="particle particle-red" filter="url(#cdh-glow-dark)">
          <animateMotion dur="1.6s" repeatCount="indefinite" keyPoints="0;1" keyTimes="0;1" calcMode="linear"><mpath href="#cdh-path-store-photo" /></animateMotion>
        </circle>

        <!-- Right-side: 封装ZTDO particle -->
        <circle r="3.5" class="particle particle-green" filter="url(#cdh-glow-green)">
          <animateMotion dur="1.4s"repeatCount="indefinite" keyPoints="0;1" keyTimes="0;1" calcMode="linear"><mpath href="#cdh-path-dhp-ztdo" /></animateMotion>
        </circle>
      </g>

      <!-- ===== FLOW STEP BADGES ===== -->
      <g class="flow-layer">
        <g v-for="step in flowSteps" :key="step.id"
          class="flow-step"
          :transform="`translate(${step.x}, ${step.y})`">
          <circle cx="0" cy="0" r="13" :class="['step-badge', `step-badge--${step.color}`]" filter="url(#cdh-shadow-badge)" />
          <text x="0" y="4.5" class="step-number">{{ step.id }}</text>
          <rect :x="(step.labelX || 18) - 3" y="-9" :width="step.labelW || 72" height="18" rx="4" class="step-label-bg" />
          <text :x="step.labelX || 18" :y="4" class="step-label">{{ step.label }}</text>
        </g>
      </g>
    </svg>

    <!-- Info Card -->
    <div class="info-card-container">
      <transition name="slide">
        <div v-if="active && info[active]" class="info-card" :style="{ borderLeftColor: info[active].color }">
          <div class="info-card-header">
            <div class="info-icon" :style="{ background: info[active].color }">
              {{ info[active].icon }}
            </div>
            <div>
              <h4 class="info-title">{{ info[active].title }}</h4>
              <span class="info-badge">{{ info[active].badge }}</span>
            </div>
          </div>
          <p class="info-desc">{{ info[active].desc }}</p>
          <div v-if="info[active].features" class="info-features">
            <span v-for="f in info[active].features" :key="f" class="info-feature">{{ f }}</span>
          </div>
        </div>
      </transition>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const active = ref<string | null>(null)

const flowSteps = [
  { id: 1, x: 55,  y: 538,  label: '发起AI任务',    color: 'blue',  labelW: 78 },
  { id: 2, x: 140, y: 700,  label: '调度任务',      color: 'blue',  labelW: 60 },
  { id: 3, x: 350, y: 362,  label: '请求远程证明',   color: 'green', labelW: 84 },
  { id: 4, x: 430, y: 340,  label: '返回证明结果',   color: 'green', labelW: 84 },
  { id: 5, x: 305, y: 340,  label: '请求授权',      color: 'green', labelW: 60, labelX: -66 },
  { id: 6, x: 495, y: 400,  label: '返回授权',      color: 'green', labelW: 60 },
  { id: 7, x: 565, y: 470,  label: '获取加密ZTDO',   color: 'red',   labelW: 90, labelX: -96 },
  { id: 8, x: 322, y: 498,  label: 'TEE内解密',     color: 'teal',  labelW: 72, labelX: -78 },
  { id: 9, x: 322, y: 652,  label: '供AI使用',      color: 'teal',  labelW: 66, labelX: -72 },
]

interface Info { title: string; badge: string; desc: string; color: string; icon: string; features?: string[] }

const info: Record<string, Info> = {
  user: {
    title: '用户',
    badge: '任务发起方',
    desc: 'AI任务的发起主体。用户通过OpenClaw发起AI任务请求，任务在可信执行环境中安全执行，用户只接收最终的AI处理结果，敏感数据全程不暴露。',
    color: '#3b82f6',
    icon: '\u{1F464}',
    features: ['发起AI任务', '接收处理结果', '数据零接触'],
  },
  openclaw: {
    title: 'OpenClaw AI任务调度',
    badge: '调度系统',
    desc: 'OpenClaw负责接收用户的AI任务请求并进行调度，将任务安全分配给可信执行环境中的AI任务引擎执行。',
    color: '#3b82f6',
    icon: '\u{1F916}',
    features: ['任务编排', '安全调度', '多模型支持', 'TEE分发'],
  },
  aitask: {
    title: 'AI 任务引擎',
    badge: '计算核心',
    desc: 'OpenClaw Worker，执行AI计算的核心组件。使用TEE内解密后的敏感数据进行AI推理计算，完成用户请求的AI任务。',
    color: '#ca8a04',
    icon: '\u{1F9E0}',
    features: ['AI推理', 'OpenClaw Worker', '机密计算', '内存安全'],
  },
  nhpserver: {
    title: 'NHP服务器',
    badge: 'NHP核心组件',
    desc: '中央认证授权服务器，处理NHP代理的远程证明请求和数据访问授权请求。验证TEE环境完整性后，授权NHP-DB向TEE内的NHP代理提供加密数据。',
    color: '#16a34a',
    icon: '\u{1F5A5}',
    features: ['认证授权', 'Noise协议', '集群部署', '策略管理'],
  },
  attestation: {
    title: '远程证明服务',
    badge: 'NHP组件',
    desc: '验证可信执行环境(TEE)的完整性和安全性。通过远程证明协议确认TEE环境未被篡改，确保敏感数据只在可信环境中处理。',
    color: '#16a34a',
    icon: '\u{1F6E1}',
    features: ['TEE验证', '完整性证明', '安全审计'],
  },
  dynkey: {
    title: '动态密钥访问对象',
    badge: 'NHP安全机制',
    desc: '采用"一用一密"的动态密钥机制。每次数据访问使用不同的加密密钥，防止密钥泄露和重放攻击，确保数据访问的安全性。',
    color: '#16a34a',
    icon: '\u{1F511}',
    features: ['一用一密', '动态生成', '防重放', '密钥隔离'],
  },
  agent: {
    title: 'NHP代理',
    badge: 'TEE内组件',
    desc: '运行在可信执行环境内的NHP代理。负责向NHP服务器发起远程证明和授权请求，从NHP-DB获取加密的ZTDO数据，并在TEE内使用动态密钥安全解密。',
    color: '#16a34a',
    icon: '\u{1F512}',
    features: ['远程证明模块', 'TEE内运行', 'ZTDO解密', '动态密钥'],
  },
  decrypted: {
    title: '解密后敏感数据',
    badge: 'TEE内存数据',
    desc: '在TEE内解密后的敏感数据仅存在于安全内存中。数据不会写入磁盘或泄露到TEE外部，确保全程数据安全。AI任务引擎从内存中直接使用这些数据。',
    color: '#0d9488',
    icon: '\u{1F4C4}',
    features: ['仅存内存', '不出TEE', '即用即销', '安全隔离'],
  },
  nhpdb: {
    title: 'DHP数据经纪件',
    badge: 'DHP组件',
    desc: '负责接收、加密、存储和管理用户敏感数据。将用户的文档、邮件、照片等加密后封装为零信任数据对象(ZTDO)，仅在NHP服务器授权后向TEE内的NHP代理提供数据。',
    color: '#dc2626',
    icon: '\u{1F4BE}',
    features: ['数据加密', '安全存储', 'ZTDO封装', '授权分发'],
  },
  ztdo: {
    title: '零信任数据对象 ZTDO',
    badge: 'DHP数据格式',
    desc: '敏感数据的加密封装格式，采用一文一钥机制。每个数据对象使用独立的加密密钥，只有通过NHP授权获得动态密钥后，才能在TEE环境中解密使用。',
    color: '#16a34a',
    icon: '\u{1F4E6}',
    features: ['一文一钥', '零信任', '加密封装', '授权解密'],
  },
  datasource: {
    title: '用户敏感数据源',
    badge: '原始数据',
    desc: '用户的原始敏感数据，包括文档、邮件、照片等。这些数据首先被加密，然后存储到DHP数据经纪件中，以零信任数据对象(ZTDO)的形式安全管理。',
    color: '#dc2626',
    icon: '\u{1F4C1}',
    features: ['文档', '邮件', '照片', '端到端加密'],
  },
}
</script>

<style scoped>
/* ===== DESIGN TOKENS (Light) ===== */
.claw-dhp-wrapper {
  --node-fill-from: #ffffff; --node-fill-to: #f8fafc;
  --node-stroke: #e2e8f0; --node-stroke-hover: #3b82f6;
  --text-primary: #1e293b; --text-secondary: #475569; --text-muted: #94a3b8;
  --surface: #ffffff; --surface-border: #e2e8f0;
  --shadow-opacity: 0.08; --shadow-opacity-md: 0.12;
  --step-label-bg: rgba(255,255,255,0.90);

  /* Zone colors */
  --zone-red-border: #dc2626; --zone-red-bg: rgba(220,38,38,0.04);
  --zone-blue-border: #3b82f6; --zone-blue-bg: rgba(59,130,246,0.04);
  --tee-border: #f59e0b; --tee-bg: rgba(245,158,11,0.06);

  /* Box colors */
  --box-plain-fill: #fff; --box-plain-stroke: #d1d5db;
  --box-pink-fill: #fce7f3; --box-pink-stroke: #f9a8d4;
  --box-blue-fill: #fff; --box-blue-stroke: #3b82f6;
  --box-agent-fill: #16a34a; --box-agent-stroke: #15803d;
  --box-decrypted-fill: #dcfce7; --box-decrypted-stroke: #86efac;
  --box-ztdo-fill: #dcfce7; --box-ztdo-stroke: #16a34a;
  --box-ai-fill: #fef9c3; --box-ai-stroke: #fbbf24;
  --box-white-fill: #fff; --box-white-stroke: #d1d5db;

  /* Arrow colors */
  --arrow-blue: #3b82f6; --arrow-green: #16a34a; --arrow-teal: #0d9488; --arrow-dark: #475569;
  --arrow-red: #dc2626; --arrow-store: #16a34a;
  --conn-blue: #3b82f6; --conn-nhp: #16a34a; --conn-data: #0d9488; --conn-task: #475569;

  /* Text colors */
  --title-color: #0f172a;
  --zone-title-red: #dc2626; --zone-title-blue: #1e40af;
  --encrypt-color: #dc2626; --store-color: #16a34a;
  --provide-key-color: #475569;
  --tee-title-color: #b45309;
  --agent-text: #fff; --agent-text-sub: rgba(255,255,255,0.8);
  --decrypted-text: #166534; --decrypted-text-sub: #15803d;
  --ztdo-title: #166534; --ztdo-sub: #15803d;
  --ai-text: #854d0e; --ai-text-sub: #a16207;

  position: relative; max-width: 1000px; margin: 1.5rem auto;
  font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
}

/* ===== DESIGN TOKENS (Dark) ===== */
:root.dark .claw-dhp-wrapper {
  --node-fill-from: #1e293b; --node-fill-to: #1e293b;
  --node-stroke: #334155; --node-stroke-hover: #60a5fa;
  --text-primary: #f1f5f9; --text-secondary: #94a3b8; --text-muted: #64748b;
  --surface: #0f172a; --surface-border: #1e293b;
  --shadow-opacity: 0.3; --shadow-opacity-md: 0.35;
  --step-label-bg: rgba(15,23,42,0.90);

  --zone-red-border: #f87171; --zone-red-bg: rgba(239,68,68,0.06);
  --zone-blue-border: #60a5fa; --zone-blue-bg: rgba(59,130,246,0.06);
  --tee-border: #fbbf24; --tee-bg: rgba(245,158,11,0.08);

  --box-plain-fill: #1e293b; --box-plain-stroke: #475569;
  --box-pink-fill: rgba(244,114,182,0.12); --box-pink-stroke: rgba(244,114,182,0.3);
  --box-blue-fill: #1e293b; --box-blue-stroke: #60a5fa;
  --box-agent-fill: #166534; --box-agent-stroke: #15803d;
  --box-decrypted-fill: rgba(22,163,74,0.12); --box-decrypted-stroke: rgba(74,222,128,0.3);
  --box-ztdo-fill: rgba(22,163,74,0.12); --box-ztdo-stroke: #4ade80;
  --box-ai-fill: rgba(202,138,4,0.15); --box-ai-stroke: rgba(251,191,36,0.4);
  --box-white-fill: #1e293b; --box-white-stroke: #475569;

  --arrow-blue: #60a5fa; --arrow-green: #4ade80; --arrow-teal: #2dd4bf; --arrow-dark: #94a3b8;
  --arrow-red: #f87171; --arrow-store: #4ade80;
  --conn-blue: #60a5fa; --conn-nhp: #4ade80; --conn-data: #2dd4bf; --conn-task: #94a3b8;

  --title-color: #f1f5f9;
  --zone-title-red: #f87171; --zone-title-blue: #60a5fa;
  --encrypt-color: #f87171; --store-color: #4ade80;
  --provide-key-color: #94a3b8;
  --tee-title-color: #fbbf24;
  --agent-text: #fff; --agent-text-sub: rgba(255,255,255,0.7);
  --decrypted-text: #4ade80; --decrypted-text-sub: #86efac;
  --ztdo-title: #4ade80; --ztdo-sub: #86efac;
  --ai-text: #fbbf24; --ai-text-sub: #fcd34d;
}

/* ===== SVG Canvas ===== */
.arch-svg { width: 100%; height: auto; border-radius: 12px; background: var(--surface); border: 1px solid var(--surface-border); }

/* ===== Title ===== */
.diagram-title { font-size: 22px; font-weight: 800; fill: var(--title-color); }

/* ===== Legend Toolbar ===== */
.arch-toolbar { display: flex; align-items: center; margin-bottom: 12px; flex-wrap: wrap; gap: 8px; }
.legend-bar { display: flex; align-items: center; gap: 14px; font-size: 12px; color: var(--text-muted); }
.legend-item { display: flex; align-items: center; gap: 6px; }
.legend-line { display: inline-block; width: 24px; height: 2.5px; border-radius: 1px; }
.legend-line.nhp { background: #16a34a; } .legend-line.data { background: #0d9488; } .legend-line.task-blue { background: #3b82f6; }
.legend-line.encrypt { background: #dc2626; border-style: dashed; height: 0; border-top: 2.5px dashed #dc2626; }
.legend-line.data-red { background: #dc2626; }
.legend-sep { color: var(--surface-border); font-size: 14px; margin: 0 2px; }
.legend-dot.req { display: inline-block; width: 8px; height: 8px; border-radius: 50%; background: #16a34a; box-shadow: 0 0 4px rgba(22,163,74,0.5); }
.legend-ring { display: inline-block; width: 8px; height: 8px; border-radius: 50%; border: 2px solid #16a34a; }

/* ===== Zone Borders ===== */
.zone-border-red { fill: var(--zone-red-bg); stroke: var(--zone-red-border); stroke-width: 2; }
.zone-border-blue { fill: var(--zone-blue-bg); stroke: var(--zone-blue-border); stroke-width: 2; }
.zone-title-inner { font-size: 13px; font-weight: 700; text-anchor: middle; }
.zone-title-red { fill: var(--zone-title-red); }
.zone-title-blue { fill: var(--zone-title-blue); }

/* ===== TEE ===== */
.tee-border { fill: var(--tee-bg); stroke: var(--tee-border); stroke-width: 3.5; }
.tee-title { font-size: 14px; font-weight: 700; fill: var(--tee-title-color); text-anchor: middle; letter-spacing: 0.5px; }

/* ===== Box Styles ===== */
.box-plain { fill: var(--box-plain-fill); stroke: var(--box-plain-stroke); stroke-width: 1.2; }
.box-pink { fill: var(--box-pink-fill); stroke: var(--box-pink-stroke); stroke-width: 1.2; }
.box-blue-border { fill: var(--box-blue-fill); stroke: var(--box-blue-stroke); stroke-width: 2; }
.box-white { fill: var(--box-white-fill); stroke: var(--box-white-stroke); stroke-width: 1.2; }
.box-agent { fill: var(--box-agent-fill); stroke: var(--box-agent-stroke); stroke-width: 1.5; }
.box-decrypted { fill: var(--box-decrypted-fill); stroke: var(--box-decrypted-stroke); stroke-width: 1.5; }
.box-ztdo { fill: var(--box-ztdo-fill); stroke: var(--box-ztdo-stroke); stroke-width: 2; }
.box-ai-engine { fill: var(--box-ai-fill); stroke: var(--box-ai-stroke); stroke-width: 1.5; }

/* ===== Box Text ===== */
.box-text { font-size: 14px; font-weight: 700; fill: var(--text-primary); text-anchor: middle; }
.box-text-sm { font-size: 13px; font-weight: 600; fill: var(--text-primary); text-anchor: middle; }
.box-text-sub { font-size: 11px; fill: var(--text-secondary); text-anchor: middle; }
.box-text-white { font-size: 14px; font-weight: 700; fill: var(--agent-text); text-anchor: middle; }
.box-text-white-sub { font-size: 10px; fill: var(--agent-text-sub); text-anchor: middle; }
.box-text-green { font-size: 13px; font-weight: 600; fill: var(--decrypted-text); text-anchor: middle; }
.box-text-green-sub { font-size: 11px; fill: var(--decrypted-text-sub); text-anchor: middle; }
.box-text-ai { font-size: 16px; font-weight: 700; fill: var(--ai-text); text-anchor: middle; }
.box-text-ai-sub { font-size: 11px; fill: var(--ai-text-sub); text-anchor: middle; }
.ztdo-title { font-size: 14px; font-weight: 700; fill: var(--ztdo-title); text-anchor: middle; }
.ztdo-subtitle { font-size: 13px; font-weight: 600; fill: var(--ztdo-sub); text-anchor: middle; letter-spacing: 1px; }
.ztdo-label-bg { fill: var(--surface); stroke: none; }

/* ===== Labels ===== */
.encrypt-label { font-size: 11px; font-weight: 600; fill: var(--encrypt-color); text-anchor: middle; }
.store-label { font-size: 11px; font-weight: 600; fill: var(--store-color); text-anchor: middle; }
.store-label-red { font-size: 11px; font-weight: 600; fill: var(--arrow-red); text-anchor: middle; }
.arrow-red-dashed { stroke: var(--arrow-red); stroke-width: 1.5; stroke-dasharray: 4,3; }
.label-provide-key { font-size: 10px; fill: var(--provide-key-color); text-anchor: middle; }
.label-blue { font-size: 10px; fill: var(--conn-blue); text-anchor: middle; font-weight: 600; }

/* ===== Static Arrows ===== */
.arrow-encrypt { stroke: var(--arrow-red); stroke-width: 1.5; }
.arrow-store { stroke: var(--arrow-store); stroke-width: 1.5; }
.arrow-provide-key { stroke: var(--provide-key-color); stroke-width: 1.5; }

/* ===== Node Hover ===== */
.node { cursor: pointer; }
.node:hover .box-plain, .node.active .box-plain,
.node:hover .box-white, .node.active .box-white,
.node:hover .box-pink, .node.active .box-pink,
.node:hover .box-blue-border, .node.active .box-blue-border,
.node:hover .box-agent, .node.active .box-agent,
.node:hover .box-decrypted, .node.active .box-decrypted,
.node:hover .box-ztdo, .node.active .box-ztdo,
.node:hover .box-ai-engine, .node.active .box-ai-engine {
  filter: url(#cdh-shadow-md); stroke-width: 2.5;
}

/* ===== Connection Lines ===== */
.conn { fill: none; stroke-width: 2; stroke-linejoin: round; stroke-linecap: round; }
.conn-blue { stroke: var(--conn-blue); }
.conn-nhp { stroke: var(--conn-nhp); }
.conn-nhp-dashed { stroke-dasharray: 6,4; }
.conn-data { stroke: var(--conn-data); }
.conn-data-dashed { stroke: var(--conn-data); stroke-dasharray: 6,4; }
.conn-red-dashed { stroke: var(--arrow-red); stroke-dasharray: 6,4; }
.conn-task { stroke: var(--conn-task); stroke-width: 2; }

/* ===== Particles ===== */
.particle { opacity: 0.9; }
.particle-blue { fill: #3b82f6; } .particle-green { fill: #16a34a; } .particle-teal { fill: #0d9488; } .particle-dark { fill: #64748b; } .particle-red { fill: #dc2626; }
.particle-green-ring { fill: none; stroke: #4ade80; stroke-width: 2; }
.particle-teal-ring { fill: none; stroke: #5eead4; stroke-width: 2; }

/* ===== Flow Step Badges ===== */
.flow-step { opacity: 1; }
.step-badge { stroke: white; stroke-width: 2.5; }
.step-badge--blue { fill: #3b82f6; }
.step-badge--green { fill: #16a34a; }
.step-badge--teal { fill: #0d9488; }
.step-badge--red { fill: #dc2626; }
.step-badge--dark { fill: #475569; }
.step-number { font-size: 11px; font-weight: 700; fill: white; text-anchor: middle; }
.step-label { font-size: 10px; font-weight: 600; fill: var(--text-primary); }
.step-label-bg { fill: var(--step-label-bg); stroke: none; }

/* ===== Info Card ===== */
.info-card-container { min-height: 60px; }
.info-card { background: var(--surface); border: 1px solid var(--surface-border); border-left: 4px solid; border-radius: 12px; padding: 16px 20px; margin-top: 12px; box-shadow: 0 4px 12px var(--info-shadow, rgba(0,0,0,0.06)); }
.info-card-header { display: flex; align-items: center; gap: 12px; margin-bottom: 10px; }
.info-icon { width: 36px; height: 36px; border-radius: 10px; display: flex; align-items: center; justify-content: center; font-size: 18px; color: white; flex-shrink: 0; }
.info-title { margin: 0; font-size: 15px; font-weight: 700; color: var(--text-primary); }
.info-badge { font-size: 11px; color: var(--text-muted); font-weight: 500; }
.info-desc { margin: 0 0 10px; font-size: 13px; color: var(--text-secondary); line-height: 1.6; }
.info-features { display: flex; gap: 6px; flex-wrap: wrap; }
.info-feature { display: inline-block; padding: 2px 10px; border-radius: 100px; font-size: 11px; font-weight: 500; background: var(--zone-blue-bg); color: var(--text-secondary); border: 1px solid var(--surface-border); }

/* ===== Transitions ===== */
.slide-enter-active, .slide-leave-active { transition: opacity 0.2s ease; }
.slide-enter-from, .slide-leave-to { opacity: 0; }

/* ===== Accessibility: Reduced Motion ===== */
@media (prefers-reduced-motion: reduce) {
  .particles-layer { display: none; }
}

/* ===== Dark mode info card shadow ===== */
:root.dark .claw-dhp-wrapper { --info-shadow: rgba(0,0,0,0.3); }

/* ===== Responsive ===== */
@media (max-width: 640px) {
  .arch-toolbar { flex-direction: column; align-items: flex-start; }
  .legend-bar { gap: 10px; font-size: 11px; }
  .info-card { padding: 12px 14px; }
  .info-features { gap: 4px; }
}
</style>
