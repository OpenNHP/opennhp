<template>
  <div class="arch-wrapper">
    <!-- Legend -->
    <div class="arch-toolbar">
      <div class="legend-bar">
        <span class="legend-item"><span class="legend-line control"></span> Control</span>
        <span class="legend-item"><span class="legend-line support"></span> Auth Service</span>
        <span class="legend-item"><span class="legend-line data"></span> Data</span>
        <span class="legend-sep">|</span>
        <span class="legend-item"><span class="legend-dot req"></span> Request</span>
        <span class="legend-item"><span class="legend-ring"></span> Response</span>
      </div>
    </div>

    <svg viewBox="-80 0 980 590" xmlns="http://www.w3.org/2000/svg" class="arch-svg"
      role="img" aria-label="OpenNHP Zero Trust architecture diagram showing the 7-step protocol flow between Agent, Server, Access Controller, Service Provider and Protected Resource">
      <defs>
        <linearGradient id="grad-service" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stop-color="var(--zone-service-from)" /><stop offset="100%" stop-color="var(--zone-service-to)" />
        </linearGradient>
        <linearGradient id="grad-external" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stop-color="var(--zone-external-from)" /><stop offset="100%" stop-color="var(--zone-external-to)" />
        </linearGradient>
        <linearGradient id="grad-node" x1="0" y1="0" x2="1" y2="1">
          <stop offset="0%" stop-color="var(--node-fill-from)" /><stop offset="100%" stop-color="var(--node-fill-to)" />
        </linearGradient>

        <filter id="shadow-sm" x="-4%" y="-4%" width="108%" height="112%">
          <feDropShadow dx="0" dy="1" stdDeviation="2" flood-opacity="var(--shadow-opacity, 0.08)" />
        </filter>
        <filter id="shadow-md" x="-6%" y="-6%" width="112%" height="120%">
          <feDropShadow dx="0" dy="2" stdDeviation="4" flood-opacity="var(--shadow-opacity-md, 0.12)" />
        </filter>
        <filter id="shadow-badge" x="-30%" y="-30%" width="160%" height="160%">
          <feDropShadow dx="0" dy="1" stdDeviation="3" flood-opacity="0.25" />
        </filter>
        <filter id="glow-blue" x="-20%" y="-20%" width="140%" height="140%">
          <feGaussianBlur stdDeviation="4" result="blur" />
          <feFlood flood-color="#3b82f6" flood-opacity="0.4" result="color" />
          <feComposite in="color" in2="blur" operator="in" result="glow" />
          <feMerge><feMergeNode in="glow" /><feMergeNode in="SourceGraphic" /></feMerge>
        </filter>
        <!-- Particle glow filters -->
        <filter id="particle-glow-blue" x="-100%" y="-100%" width="300%" height="300%">
          <feGaussianBlur stdDeviation="3" result="blur" />
          <feFlood flood-color="#3b82f6" flood-opacity="0.5" result="color" />
          <feComposite in="color" in2="blur" operator="in" result="glow" />
          <feMerge><feMergeNode in="glow" /><feMergeNode in="SourceGraphic" /></feMerge>
        </filter>
        <filter id="particle-glow-orange" x="-100%" y="-100%" width="300%" height="300%">
          <feGaussianBlur stdDeviation="3" result="blur" />
          <feFlood flood-color="#f59e0b" flood-opacity="0.5" result="color" />
          <feComposite in="color" in2="blur" operator="in" result="glow" />
          <feMerge><feMergeNode in="glow" /><feMergeNode in="SourceGraphic" /></feMerge>
        </filter>
        <filter id="particle-glow-green" x="-100%" y="-100%" width="300%" height="300%">
          <feGaussianBlur stdDeviation="3" result="blur" />
          <feFlood flood-color="#10b981" flood-opacity="0.5" result="color" />
          <feComposite in="color" in2="blur" operator="in" result="glow" />
          <feMerge><feMergeNode in="glow" /><feMergeNode in="SourceGraphic" /></feMerge>
        </filter>

        <!-- Arrow markers (end-only for cleaner look) -->
        <marker id="arr-blue" viewBox="0 0 8 6" refX="7" refY="3" markerWidth="8" markerHeight="6" orient="auto-start-reverse">
          <path d="M0,0.5 L7,3 L0,5.5" fill="none" stroke="#3b82f6" stroke-width="1.2" stroke-linejoin="round" /></marker>
        <marker id="arr-orange" viewBox="0 0 8 6" refX="7" refY="3" markerWidth="8" markerHeight="6" orient="auto-start-reverse">
          <path d="M0,0.5 L7,3 L0,5.5" fill="none" stroke="#f59e0b" stroke-width="1.2" stroke-linejoin="round" /></marker>
        <marker id="arr-green" viewBox="0 0 8 6" refX="7" refY="3" markerWidth="8" markerHeight="6" orient="auto-start-reverse">
          <path d="M0,0.5 L7,3 L0,5.5" fill="none" stroke="#10b981" stroke-width="1.2" stroke-linejoin="round" /></marker>
        <marker id="arr-red" viewBox="0 0 8 6" refX="7" refY="3" markerWidth="8" markerHeight="6" orient="auto-start-reverse">
          <path d="M0,0.5 L7,3 L0,5.5" fill="none" stroke="#ef4444" stroke-width="1.2" stroke-linejoin="round" /></marker>

        <!-- ClipPaths for split-zone coloring -->
        <clipPath id="clip-requestor"><rect x="20" y="235" width="240" height="330" rx="12" /></clipPath>
        <clipPath id="clip-provider"><rect x="280" y="235" width="600" height="330" rx="12" /></clipPath>

        <!-- Single polyline paths (orthogonal routing) -->
        <path id="path-agent-server" d="M 228,335 L 274,335 L 274,135 L 340,135" />
        <path id="path-server-asp"   d="M 510,135 L 640,135" />
        <path id="path-server-ac"    d="M 425,190 L 425,275" />
        <path id="path-ac-decision"  d="M 425,395 L 425,448" />
        <path id="path-user-allow"   d="M 160,467 L 340,467" />
        <path id="path-allow-resource" d="M 510,467 L 630,467" />
        <path id="path-deny-resource"  d="M 510,500 L 630,500" />
      </defs>

      <!-- ===== ZONE: NHP SERVICE ===== -->
      <g class="zone">
        <rect x="280" y="16" width="280" height="204" rx="12" class="zone-bg" fill="url(#grad-service)" />
        <text x="420" y="42" class="zone-title">NHP Service</text>
        <!-- Server: cx=425, cy=135 -->
        <g class="node"
          :class="{ active: active === 'server' }"
          @mouseenter="active = 'server'" @mouseleave="active = null" @click="handleClick('server')">
          <rect x="340" y="80" width="170" height="110" rx="16" class="node-card" filter="url(#shadow-sm)" />
          <g transform="translate(398,92)">
            <rect x="0" y="0" width="54" height="12" rx="3" class="icon-rack" />
            <rect x="0" y="15" width="54" height="12" rx="3" class="icon-rack" />
            <rect x="0" y="30" width="54" height="12" rx="3" class="icon-rack" />
            <circle cx="42" cy="6" r="2.5" class="icon-led" />
            <circle cx="42" cy="21" r="2.5" class="icon-led" />
            <circle cx="42" cy="36" r="2.5" class="icon-led" />
          </g>
          <text x="425" y="155" class="node-name">NHP-Server</text>
          <text x="425" y="171" class="node-role">Authentication &amp; Routing</text>
        </g>
      </g>

      <!-- ===== ZONE: AUTHORIZATION (Third Party) ===== -->
      <g class="zone">
        <rect x="610" y="16" width="270" height="204" rx="12" class="zone-bg zone-external" />
        <text x="720" y="42" class="zone-title">Authorization</text>
        <rect x="790" y="8" width="82" height="20" rx="10" class="badge-3rd" />
        <text x="831" y="22" class="badge-3rd-text">3rd Party</text>
        <!-- ASP: cx=730, cy=135 -->
        <g class="node"
          :class="{ active: active === 'asp' }"
          @mouseenter="active = 'asp'" @mouseleave="active = null">
          <rect x="640" y="80" width="180" height="110" rx="16" class="node-card" filter="url(#shadow-sm)" />
          <g transform="translate(703,92)">
            <path d="M27,2 L46,12 L46,26 C46,38 36,46 27,50 C18,46 8,38 8,26 L8,12 Z" class="icon-shield" />
            <path d="M20,26 L25,31 L34,20" fill="none" stroke="var(--icon-check)" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" />
          </g>
          <text x="730" y="155" class="node-name">Service Provider</text>
          <text x="730" y="171" class="node-role">Policy &amp; Identity</text>
        </g>
      </g>

      <!-- ===== ZONE: RESOURCE REQUESTOR (split at y=400) ===== -->
      <g class="zone">
        <rect x="20" y="235" width="240" height="187" class="zone-fill-requestor-upper" clip-path="url(#clip-requestor)" />
        <rect x="20" y="422" width="240" height="143" class="zone-fill-requestor-lower" clip-path="url(#clip-requestor)" />
        <rect x="20" y="235" width="240" height="330" rx="12" fill="none" class="zone-border" />
        <text x="140" y="261" class="zone-title">Resource Requestor</text>
        <!-- Agent: cx=135, cy=335 -->
        <g class="node"
          :class="{ active: active === 'agent' }"
          @mouseenter="active = 'agent'" @mouseleave="active = null" @click="handleClick('agent')">
          <rect x="42" y="275" width="186" height="120" rx="16" class="node-card" filter="url(#shadow-sm)" />
          <g transform="translate(103,293)">
            <rect x="0" y="0" width="44" height="30" rx="4" class="icon-screen" />
            <rect x="-6" y="30" width="56" height="4" rx="2" class="icon-base" />
            <path d="M14,12 L22,18 L14,24Z" class="icon-play" />
          </g>
          <text x="135" y="360" class="node-name">NHP-Agent</text>
          <text x="135" y="376" class="node-role">Client / SDK</text>
        </g>
        <!-- User (Data Plane) — aligned with Allow Access row -->
        <g class="user-actor" transform="translate(75,440)">
          <circle cx="35" cy="10" r="10" class="icon-user-head" />
          <path d="M15,40 C15,27 55,27 55,40" class="icon-user-body" />
          <text x="35" y="60" class="actor-label">User</text>
        </g>
      </g>

      <!-- ===== ZONE: RESOURCE PROVIDER (split at y=400) ===== -->
      <g class="zone">
        <rect x="280" y="235" width="600" height="187" class="zone-fill-provider-upper" clip-path="url(#clip-provider)" />
        <rect x="280" y="422" width="600" height="143" class="zone-fill-provider-lower" clip-path="url(#clip-provider)" />
        <rect x="280" y="235" width="600" height="330" rx="12" fill="none" class="zone-border" />
        <text x="580" y="261" class="zone-title">Resource Provider</text>

        <!-- AC: cx=425, cy=335 -->
        <g class="node"
          :class="{ active: active === 'ac' }"
          @mouseenter="active = 'ac'" @mouseleave="active = null" @click="handleClick('ac')">
          <rect x="340" y="275" width="170" height="120" rx="16" class="node-card" filter="url(#shadow-sm)" />
          <g transform="translate(396,290)">
            <rect x="0" y="0" width="18" height="42" rx="2" class="icon-wall" />
            <rect x="22" y="0" width="18" height="42" rx="2" class="icon-wall" />
            <rect x="44" y="0" width="18" height="42" rx="2" class="icon-wall" />
            <line x1="9" y1="14" x2="9" y2="28" class="icon-wall-line" />
            <line x1="31" y1="14" x2="31" y2="28" class="icon-wall-line" />
            <line x1="53" y1="14" x2="53" y2="28" class="icon-wall-line" />
          </g>
          <text x="425" y="360" class="node-name">NHP-AC</text>
          <text x="425" y="376" class="node-role">Access Controller</text>
        </g>

        <!-- Decision Panel (Data Plane): top=448 -->
        <g class="decision-panel" transform="translate(340,448)">
          <rect x="0" y="0" width="170" height="70" rx="10" class="decision-bg" filter="url(#shadow-sm)" />
          <g class="decision-row allow">
            <rect x="4" y="4" width="162" height="31" rx="8" class="decision-allow-bg" />
            <circle cx="22" cy="19" r="7" class="decision-icon-allow" />
            <path d="M18,19 L21,22 L26,16" fill="none" stroke="white" stroke-width="1.5" stroke-linecap="round" />
            <text x="38" y="23" class="decision-text allow-text">Allow Access</text>
          </g>
          <g class="decision-row decline" transform="translate(0,35)">
            <rect x="4" y="0" width="162" height="31" rx="8" class="decision-decline-bg" />
            <circle cx="22" cy="15" r="7" class="decision-icon-decline" />
            <line x1="18" y1="11" x2="26" y2="19" stroke="white" stroke-width="1.5" stroke-linecap="round" />
            <line x1="26" y1="11" x2="18" y2="19" stroke="white" stroke-width="1.5" stroke-linecap="round" />
            <text x="38" y="19" class="decision-text decline-text">Deny Access</text>
          </g>
        </g>

        <!-- Protected Resource (Data Plane) -->
        <g class="node resource-node"
          :class="{ active: active === 'resource' }"
          @mouseenter="active = 'resource'" @mouseleave="active = null">
          <rect x="630" y="443" width="200" height="100" rx="16" class="node-card resource-card" filter="url(#shadow-sm)" />
          <g transform="translate(690,451)">
            <ellipse cx="30" cy="6" rx="24" ry="8" class="icon-db-top" />
            <path d="M6,6 L6,34 C6,40 16,44 30,44 C44,44 54,40 54,34 L54,6" class="icon-db-body" />
            <ellipse cx="30" cy="34" rx="24" ry="8" class="icon-db-bottom" />
            <ellipse cx="30" cy="20" rx="24" ry="8" class="icon-db-mid" />
          </g>
          <text x="730" y="513" class="node-name">Protected Resource</text>
          <text x="730" y="529" class="node-role">API / Server / Gateway</text>
          <!-- Lock badge -->
          <g transform="translate(802,449)">
            <circle cx="10" cy="10" r="10" class="badge-lock-bg" />
            <rect x="5" y="9" width="10" height="8" rx="2" class="badge-lock-body" />
            <path d="M7,9 L7,6 C7,3 13,3 13,6 L13,9" fill="none" class="badge-lock-shackle" stroke-width="1.5" />
          </g>
        </g>
      </g>

      <!-- ===== PLANE DIVIDER (between NHP-AC and Decision Panel) ===== -->
      <line x1="-80" y1="422" x2="880" y2="422" class="plane-line" />
      <g class="plane-labels">
        <text x="-40" y="408" class="plane-text" text-anchor="middle">CONTROL</text>
        <text x="-40" y="419" class="plane-text" text-anchor="middle">PLANE</text>
        <text x="-40" y="437" class="plane-text" text-anchor="middle">DATA</text>
        <text x="-40" y="448" class="plane-text" text-anchor="middle">PLANE</text>
      </g>

      <!-- ===== CONNECTIONS ===== -->
      <g class="connections">
        <use href="#path-agent-server" class="conn conn-control" marker-end="url(#arr-blue)" />
        <use href="#path-server-asp" class="conn conn-support" marker-end="url(#arr-orange)" />
        <use href="#path-server-ac" class="conn conn-control" marker-end="url(#arr-blue)" />
        <use href="#path-ac-decision" class="conn conn-control" marker-end="url(#arr-blue)" />
        <use href="#path-user-allow" class="conn conn-data" marker-end="url(#arr-green)" />
        <use href="#path-allow-resource" class="conn conn-data" marker-end="url(#arr-green)" />
        <!-- Deny ✕ Resource -->
        <use href="#path-deny-resource" class="conn conn-deny" marker-end="url(#arr-red)" />
        <g transform="translate(570,500)">
          <circle cx="0" cy="0" r="10" class="block-circle" />
          <line x1="-5" y1="-5" x2="5" y2="5" stroke="white" stroke-width="2" stroke-linecap="round" />
          <line x1="5" y1="-5" x2="-5" y2="5" stroke="white" stroke-width="2" stroke-linecap="round" />
        </g>
      </g>

      <!-- ===== PARTICLES ===== -->
      <g class="particles-layer">
        <!-- Agent → Server (request) -->
        <circle r="5" class="particle particle-req particle-blue" filter="url(#particle-glow-blue)">
          <animateMotion dur="2.6s" repeatCount="indefinite" keyPoints="0;1" keyTimes="0;1" calcMode="linear"><mpath href="#path-agent-server" /></animateMotion>
        </circle>
        <!-- Server → Agent (response) -->
        <circle r="4" class="particle particle-resp particle-blue-ring" filter="url(#particle-glow-blue)">
          <animateMotion dur="3.2s" begin="1.3s" repeatCount="indefinite" keyPoints="1;0" keyTimes="0;1" calcMode="linear"><mpath href="#path-agent-server" /></animateMotion>
        </circle>

        <!-- Server → ASP -->
        <circle r="4.5" class="particle particle-req particle-orange" filter="url(#particle-glow-orange)">
          <animateMotion dur="1.6s" repeatCount="indefinite" keyPoints="0;1" keyTimes="0;1" calcMode="linear"><mpath href="#path-server-asp" /></animateMotion>
        </circle>
        <!-- ASP → Server -->
        <circle r="3.5" class="particle particle-resp particle-orange-ring" filter="url(#particle-glow-orange)">
          <animateMotion dur="2s" begin="0.8s" repeatCount="indefinite" keyPoints="1;0" keyTimes="0;1" calcMode="linear"><mpath href="#path-server-asp" /></animateMotion>
        </circle>

        <!-- Server → AC -->
        <circle r="5" class="particle particle-req particle-blue" filter="url(#particle-glow-blue)">
          <animateMotion dur="1.6s" repeatCount="indefinite" keyPoints="0;1" keyTimes="0;1" calcMode="linear"><mpath href="#path-server-ac" /></animateMotion>
        </circle>
        <!-- AC → Server -->
        <circle r="4" class="particle particle-resp particle-blue-ring" filter="url(#particle-glow-blue)">
          <animateMotion dur="2s" begin="0.8s" repeatCount="indefinite" keyPoints="1;0" keyTimes="0;1" calcMode="linear"><mpath href="#path-server-ac" /></animateMotion>
        </circle>

        <!-- AC → Decision -->
        <circle r="4" class="particle particle-req particle-blue" filter="url(#particle-glow-blue)">
          <animateMotion dur="1.8s" repeatCount="indefinite" keyPoints="0;1" keyTimes="0;1" calcMode="linear"><mpath href="#path-ac-decision" /></animateMotion>
        </circle>
        <!-- Decision → AC -->
        <circle r="3" class="particle particle-resp particle-blue-ring" filter="url(#particle-glow-blue)">
          <animateMotion dur="2.2s" begin="0.9s" repeatCount="indefinite" keyPoints="1;0" keyTimes="0;1" calcMode="linear"><mpath href="#path-ac-decision" /></animateMotion>
        </circle>

        <!-- User → Allow -->
        <circle r="5" class="particle particle-req particle-green" filter="url(#particle-glow-green)">
          <animateMotion dur="2s" repeatCount="indefinite" keyPoints="0;1" keyTimes="0;1" calcMode="linear"><mpath href="#path-user-allow" /></animateMotion>
        </circle>
        <!-- Allow → User -->
        <circle r="4" class="particle particle-resp particle-green-ring" filter="url(#particle-glow-green)">
          <animateMotion dur="2.4s" begin="1s" repeatCount="indefinite" keyPoints="1;0" keyTimes="0;1" calcMode="linear"><mpath href="#path-user-allow" /></animateMotion>
        </circle>

        <!-- Allow → Resource -->
        <circle r="4.5" class="particle particle-req particle-green" filter="url(#particle-glow-green)">
          <animateMotion dur="1.6s" repeatCount="indefinite" keyPoints="0;1" keyTimes="0;1" calcMode="linear"><mpath href="#path-allow-resource" /></animateMotion>
        </circle>
        <!-- Resource → Allow -->
        <circle r="3.5" class="particle particle-resp particle-green-ring" filter="url(#particle-glow-green)">
          <animateMotion dur="2s" begin="0.8s" repeatCount="indefinite" keyPoints="1;0" keyTimes="0;1" calcMode="linear"><mpath href="#path-allow-resource" /></animateMotion>
        </circle>
      </g>

      <!-- ===== FLOW STEP BADGES (always visible, color-coded) ===== -->
      <g class="flow-layer">
        <g v-for="step in flowSteps" :key="step.id"
          class="flow-step"
          :transform="`translate(${step.x}, ${step.y})`">
          <circle cx="0" cy="0" r="14" :class="['step-badge', `step-badge--${step.color}`]" filter="url(#shadow-badge)" />
          <text x="0" y="5" class="step-number">{{ step.id }}</text>
          <!-- Label with background pill -->
          <rect :x="(step.labelX || 20) - 3" y="-9" :width="step.labelW || 48" height="18" rx="4" class="step-label-bg" />
          <text :x="step.labelX || 20" :y="4" class="step-label">{{ step.label }}</text>
        </g>
      </g>
    </svg>

    <!-- Info Card -->
    <div class="info-card-container">
      <transition name="slide">
        <div v-if="active && info[active]" class="info-card" :style="{ borderLeftColor: info[active].color }">
          <div class="info-card-header">
            <div class="info-icon" :style="{ background: info[active].color }">
              <span v-html="info[active].icon"></span>
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
  { id: 1, x: 296, y: 260, label: 'Knock', color: 'blue', labelW: 50 },
  { id: 2, x: 575, y: 115, label: 'Query', color: 'orange', labelW: 46 },
  { id: 3, x: 575, y: 160, label: 'Authorize', color: 'orange', labelW: 66 },
  { id: 4, x: 450, y: 225, label: 'Open', color: 'blue', labelW: 42 },
  { id: 5, x: 400, y: 245, label: 'Confirm', labelX: -65, color: 'blue', labelW: 58 },
  { id: 6, x: 252, y: 300, label: 'ACK', labelX: -42, color: 'blue', labelW: 36 },
  { id: 7, x: 250, y: 467, label: 'Access', color: 'green', labelW: 52 },
]

interface Info { title: string; badge: string; desc: string; color: string; icon: string; features?: string[]; link?: string }

const info: Record<string, Info> = {
  server: { title: 'NHP-Server', badge: 'Core Component', desc: 'Central controller that processes and validates knock requests, interacts with the Authorization Service Provider, and orchestrates access control.', color: '#3b82f6', icon: '&#9741;', features: ['Noise Protocol', 'Key Management', 'Distributed Deploy'], link: '/code#nhp-server' },
  agent: { title: 'NHP-Agent', badge: 'Client Component', desc: 'Client-side component that initiates encrypted knock requests to gain access to protected resources. Available as standalone app, SDK, or browser plugin.', color: '#8b5cf6', icon: '&#9000;', features: ['Encrypted UDP', 'Multi-platform', 'SDK Available'], link: '/code#nhp-agent' },
  ac: { title: 'NHP-AC', badge: 'Enforcement Point', desc: 'Access Controller that enforces deny-all by default. Opens network paths only upon NHP-Server instruction, ensuring complete infrastructure hiding.', color: '#f59e0b', icon: '&#9855;', features: ['Default Deny-All', 'eBPF Support', 'Auto Expire'], link: '/code#nhp-ac' },
  asp: { title: 'Authorization Service Provider', badge: 'Third-Party External Service', desc: 'An external, independent service that validates access policies and provides actual resource addresses. Operates outside the NHP infrastructure boundary.', color: '#10b981', icon: '&#10003;', features: ['Policy Engine', 'SDP Compatible', 'RBAC / ABAC', 'Independent'] },
  resource: { title: 'Protected Resource', badge: 'Protected Target', desc: 'The target service, API, or infrastructure hidden from unauthorized access. Completely invisible on the network until NHP grants access.', color: '#6366f1', icon: '&#9881;', features: ['Zero Exposure', 'Port Hiding', 'IP Concealment'] },
}

function handleClick(id: string) {
  const item = info[id]
  if (item?.link) window.location.href = item.link
}
</script>

<style scoped>
/* ===== DESIGN TOKENS (Light) ===== */
.arch-wrapper {
  --zone-service-from: #eff6ff; --zone-service-to: #dbeafe;
  --zone-external-from: #fefce8; --zone-external-to: #fef9c3;
  --zone-requestor-upper: #f5f3ff; --zone-requestor-lower: #ddd6fe;
  --zone-provider-upper: #ecfdf5; --zone-provider-lower: #a7f3d0;
  --zone-border-color: #cbd5e1;
  --node-fill-from: #ffffff; --node-fill-to: #f8fafc;
  --node-stroke: #e2e8f0; --node-stroke-hover: #3b82f6;
  --text-primary: #1e293b; --text-secondary: #64748b; --text-muted: #94a3b8;
  --icon-rack: #cbd5e1; --icon-rack-stroke: #94a3b8; --icon-led: #10b981;
  --icon-shield: #d1fae5; --icon-shield-stroke: #10b981; --icon-check: #fff;
  --icon-wall: #fef3c7; --icon-wall-stroke: #f59e0b; --icon-wall-line: #f59e0b;
  --icon-screen: #e0e7ff; --icon-screen-stroke: #818cf8; --icon-base: #c7d2fe; --icon-play: #6366f1;
  --icon-db: #e0e7ff; --icon-db-stroke: #818cf8;
  --icon-user: #cbd5e1; --icon-user-stroke: #94a3b8;
  --badge-lock-bg: #fecaca; --badge-lock: #ef4444;
  --badge-3rd-bg: #fef3c7; --badge-3rd-stroke: #f59e0b; --badge-3rd-text: #92400e;
  --decision-bg: #ffffff; --decision-allow: #dcfce7; --decision-decline: #fee2e2;
  --decision-allow-text: #16a34a; --decision-decline-text: #ef4444;
  --plane-line: #94a3b8; --plane-text: #64748b;
  --surface: #ffffff; --surface-border: #e2e8f0;
  --block-bg: #ef4444;
  --shadow-opacity: 0.08; --shadow-opacity-md: 0.12;
  --step-label-bg: rgba(255,255,255,0.88);
  position: relative; max-width: 900px; margin: 1.5rem auto;
  font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
}

/* ===== DESIGN TOKENS (Dark) ===== */
:root.dark .arch-wrapper {
  --zone-service-from: rgba(30,58,138,0.2); --zone-service-to: rgba(30,64,175,0.12);
  --zone-external-from: rgba(245,158,11,0.14); --zone-external-to: rgba(245,158,11,0.07);
  --zone-requestor-upper: rgba(91,33,182,0.1); --zone-requestor-lower: rgba(91,33,182,0.24);
  --zone-provider-upper: rgba(6,78,59,0.1); --zone-provider-lower: rgba(5,150,105,0.24);
  --zone-border-color: #475569;
  --node-fill-from: #1e293b; --node-fill-to: #1e293b;
  --node-stroke: #334155; --node-stroke-hover: #60a5fa;
  --text-primary: #f1f5f9; --text-secondary: #94a3b8; --text-muted: #64748b;
  --icon-rack: #475569; --icon-rack-stroke: #64748b; --icon-led: #34d399;
  --icon-shield: rgba(16,185,129,0.2); --icon-shield-stroke: #34d399;
  --icon-wall: rgba(245,158,11,0.15); --icon-wall-stroke: #fbbf24; --icon-wall-line: #fbbf24;
  --icon-screen: rgba(129,140,248,0.15); --icon-screen-stroke: #a5b4fc;
  --icon-base: rgba(129,140,248,0.25); --icon-play: #a5b4fc;
  --icon-db: rgba(129,140,248,0.15); --icon-db-stroke: #a5b4fc;
  --icon-user: #475569; --icon-user-stroke: #64748b;
  --badge-lock-bg: rgba(239,68,68,0.2); --badge-lock: #f87171;
  --badge-3rd-bg: rgba(245,158,11,0.15); --badge-3rd-stroke: #fbbf24; --badge-3rd-text: #fbbf24;
  --decision-bg: #1e293b; --decision-allow: rgba(22,163,74,0.2); --decision-decline: rgba(239,68,68,0.18);
  --decision-allow-text: #4ade80; --decision-decline-text: #f87171;
  --plane-line: #475569; --plane-text: #94a3b8;
  --surface: #0f172a; --surface-border: #1e293b;
  --block-bg: #dc2626;
  --shadow-opacity: 0.3; --shadow-opacity-md: 0.35;
  --step-label-bg: rgba(15,23,42,0.88);
}

/* ===== SVG Canvas ===== */
.arch-svg { width: 100%; height: auto; border-radius: 12px; background: var(--surface); border: 1px solid var(--surface-border); }

/* ===== Legend Toolbar ===== */
.arch-toolbar { display: flex; align-items: center; margin-bottom: 12px; flex-wrap: wrap; gap: 8px; }
.legend-bar { display: flex; align-items: center; gap: 16px; font-size: 12px; color: var(--text-muted); }
.legend-item { display: flex; align-items: center; gap: 6px; }
.legend-line { display: inline-block; width: 24px; height: 2.5px; border-radius: 1px; }
.legend-line.control { background: #3b82f6; } .legend-line.support { background: #f59e0b; } .legend-line.data { background: #10b981; }
.legend-sep { color: var(--surface-border); font-size: 14px; margin: 0 2px; }
.legend-dot.req { display: inline-block; width: 8px; height: 8px; border-radius: 50%; background: #3b82f6; box-shadow: 0 0 4px rgba(59,130,246,0.5); }
.legend-ring { display: inline-block; width: 8px; height: 8px; border-radius: 50%; border: 2px solid #3b82f6; }

/* ===== Zones ===== */
.zone-bg { transition: all 0.3s ease; }
.zone-external { fill: url(#grad-external); stroke: var(--badge-3rd-stroke); stroke-width: 1; stroke-dasharray: 6,3; }
.zone-border { stroke: var(--zone-border-color); stroke-width: 1.5; }
.zone-title { font-size: 12px; font-weight: 600; fill: var(--text-secondary); text-anchor: middle; text-transform: uppercase; letter-spacing: 1.5px; }
.zone-fill-requestor-upper { fill: var(--zone-requestor-upper); } .zone-fill-requestor-lower { fill: var(--zone-requestor-lower); }
.zone-fill-provider-upper { fill: var(--zone-provider-upper); } .zone-fill-provider-lower { fill: var(--zone-provider-lower); }
.badge-3rd { fill: var(--badge-3rd-bg); stroke: var(--badge-3rd-stroke); stroke-width: 1; }
.badge-3rd-text { font-size: 9px; font-weight: 600; fill: var(--badge-3rd-text); text-anchor: middle; text-transform: uppercase; letter-spacing: 0.8px; }

/* ===== Node Cards ===== */
.node-card { fill: url(#grad-node); stroke: var(--node-stroke); stroke-width: 1.5; transition: all 0.25s ease; }
.node { cursor: pointer; }
.node:hover .node-card, .node.active .node-card { stroke: var(--node-stroke-hover); stroke-width: 2; filter: url(#shadow-md); }
.node-name { font-size: 13px; font-weight: 700; fill: var(--text-primary); text-anchor: middle; }
.node-role { font-size: 11px; fill: var(--text-secondary); text-anchor: middle; letter-spacing: 0.3px; }

/* ===== Icons ===== */
.icon-rack { fill: var(--icon-rack); stroke: var(--icon-rack-stroke); stroke-width: 1; }
.icon-led { fill: var(--icon-led); }
.icon-shield { fill: var(--icon-shield); stroke: var(--icon-shield-stroke); stroke-width: 1.5; }
.icon-wall { fill: var(--icon-wall); stroke: var(--icon-wall-stroke); stroke-width: 1; }
.icon-wall-line { stroke: var(--icon-wall-line); stroke-width: 1.5; stroke-dasharray: 2,3; }
.icon-screen { fill: var(--icon-screen); stroke: var(--icon-screen-stroke); stroke-width: 1.5; }
.icon-base { fill: var(--icon-base); } .icon-play { fill: var(--icon-play); }
.icon-db-top, .icon-db-bottom, .icon-db-mid { fill: none; stroke: var(--icon-db-stroke); stroke-width: 1.5; }
.icon-db-top { fill: var(--icon-db); } .icon-db-body { fill: var(--icon-db); stroke: var(--icon-db-stroke); stroke-width: 1.5; }
.icon-user-head { fill: var(--icon-user); stroke: var(--icon-user-stroke); stroke-width: 1.2; }
.icon-user-body { fill: none; stroke: var(--icon-user-stroke); stroke-width: 2.5; }
.actor-label { font-size: 11px; fill: var(--text-secondary); text-anchor: middle; font-weight: 600; }
.badge-lock-bg { fill: var(--badge-lock-bg); } .badge-lock-body { fill: var(--badge-lock); stroke: none; } .badge-lock-shackle { stroke: var(--badge-lock); }
.resource-card { stroke-dasharray: none; }

/* ===== Decision Panel ===== */
.decision-bg { fill: var(--decision-bg); stroke: var(--node-stroke); stroke-width: 1; }
.decision-allow-bg { fill: var(--decision-allow); stroke: none; } .decision-decline-bg { fill: var(--decision-decline); stroke: none; }
.decision-icon-allow { fill: var(--decision-allow-text); } .decision-icon-decline { fill: var(--decision-decline-text); }
.decision-text { font-size: 11px; font-weight: 600; } .allow-text { fill: var(--decision-allow-text); } .decline-text { fill: var(--decision-decline-text); }

/* ===== Plane Divider ===== */
.plane-line { stroke: var(--plane-line); stroke-width: 1.5; stroke-dasharray: 6,4; }
.plane-text { font-size: 10px; font-weight: 700; fill: var(--plane-text); letter-spacing: 1.2px; text-transform: uppercase; }

/* ===== Connection Lines ===== */
.conn { fill: none; stroke-width: 2; stroke-linejoin: round; stroke-linecap: round; }
.conn-control { stroke: #3b82f6; } .conn-support { stroke: #f59e0b; } .conn-data { stroke: #10b981; }
.conn-deny { stroke: #ef4444; stroke-dasharray: 5,4; stroke-width: 1.5; }
.block-circle { fill: var(--block-bg); stroke: white; stroke-width: 1.5; }

/* ===== Particles ===== */
.particle { opacity: 0.9; }
.particle-req { opacity: 1; }
.particle-resp { opacity: 0.75; }
.particle-blue { fill: #3b82f6; } .particle-orange { fill: #f59e0b; } .particle-green { fill: #10b981; }
.particle-blue-ring { fill: none; stroke: #60a5fa; stroke-width: 2.2; }
.particle-orange-ring { fill: none; stroke: #fbbf24; stroke-width: 2.2; }
.particle-green-ring { fill: none; stroke: #34d399; stroke-width: 2.2; }

/* ===== Flow Step Badges ===== */
.flow-step { opacity: 1; }
.step-badge { stroke: white; stroke-width: 2.5; }
.step-badge--blue { fill: #3b82f6; }
.step-badge--orange { fill: #f59e0b; }
.step-badge--green { fill: #10b981; }
.step-number { font-size: 12px; font-weight: 700; fill: white; text-anchor: middle; }
.step-label { font-size: 11px; font-weight: 600; fill: var(--text-primary); }
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
.info-feature { display: inline-block; padding: 2px 10px; border-radius: 100px; font-size: 11px; font-weight: 500; background: var(--zone-service-from); color: var(--text-secondary); border: 1px solid var(--surface-border); }

/* ===== Transitions ===== */
.slide-enter-active, .slide-leave-active { transition: opacity 0.2s ease; }
.slide-enter-from, .slide-leave-to { opacity: 0; }

/* ===== Accessibility: Reduced Motion ===== */
@media (prefers-reduced-motion: reduce) {
  .particles-layer { display: none; }
}

/* ===== Dark mode info card shadow ===== */
:root.dark .arch-wrapper { --info-shadow: rgba(0,0,0,0.3); }

/* ===== Responsive ===== */
@media (max-width: 640px) {
  .arch-toolbar { flex-direction: column; align-items: flex-start; }
  .legend-bar { gap: 10px; font-size: 11px; }
  .info-card { padding: 12px 14px; }
  .info-features { gap: 4px; }
}
</style>
