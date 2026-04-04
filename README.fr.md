[![en](https://img.shields.io/badge/lang-en-green.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.md)
[![zh-cn](https://img.shields.io/badge/lang-zh--cn-green.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.zh-cn.md)
[![de](https://img.shields.io/badge/lang-de-green.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.de.md)
[![ja](https://img.shields.io/badge/lang-ja-green.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.ja.md)
[![fr](https://img.shields.io/badge/lang-fr-green.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.fr.md)
[![es](https://img.shields.io/badge/lang-es-green.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.es.md)

![Logo OpenNHP](docs/images/logo11.png)
# OpenNHP : Protocole de Masquage de l'Infrastructure Réseau en Zéro Confiance
Un protocole réseau de zéro confiance, basé sur la cryptographie, au niveau 5 du modèle OSI, permettant de cacher votre serveur et vos données des attaquants.

![Statut de Construction](https://img.shields.io/badge/build-passing-brightgreen)
![Version](https://img.shields.io/badge/version-1.0.0-blue)
![Licence](https://img.shields.io/badge/license-Apache%202.0-green)

---

## Défi : L'IA transforme Internet en une "Forêt Sombre"

L'avancement rapide des technologies d'**IA**, en particulier les grands modèles de langage (LLM), transforme de manière significative le paysage de la cybersécurité. L'émergence de l'**exploitation autonome des vulnérabilités (AVE)** représente un bond majeur dans l'ère de l'IA, automatisant l'exploitation des vulnérabilités, comme le montre [cet article de recherche](https://arxiv.org/abs/2404.08144). Ce développement augmente de manière significative le risque pour tous les services réseau exposés, évoquant l'hypothèse de la [forêt sombre](https://fr.wikipedia.org/wiki/For%C3%AAt_sombre) sur Internet. Les outils pilotés par l'IA scannent continuellement l'environnement numérique, identifiant rapidement les faiblesses et les exploitant. Ainsi, Internet devient une **"forêt sombre"** où **la visibilité équivaut à la vulnérabilité**.

![Risques de Vulnérabilité](docs/images/Vul_Risks.png)

Selon les recherches de Gartner, les [cyberattaques pilotées par l'IA vont augmenter rapidement](https://www.gartner.com/en/newsroom/press-releases/2024-08-28-gartner-forecasts-global-information-security-spending-to-grow-15-percent-in-2025). Ce paradigme en évolution impose une réévaluation des stratégies de cybersécurité traditionnelles, avec un accent sur les défenses proactives, des mécanismes de réponse rapide, et l'adoption de technologies de masquage réseau pour protéger les infrastructures critiques.

---

## Démo rapide : Voir OpenNHP en action

Avant de plonger dans les détails d'OpenNHP, commençons par une démonstration rapide de la façon dont OpenNHP protège un serveur contre les accès non autorisés. Vous pouvez le voir en action en accédant au serveur protégé à l'adresse suivante : https://acdemo.opennhp.org.

### 1) Le serveur protégé est "invisible" aux utilisateurs non authentifiés

Par défaut, toute tentative de connexion au serveur protégé résultera en une erreur de TYPE OUT, car tous les ports sont fermés, rendant le serveur *"invisible"* et apparemment hors ligne.

![Démo OpenNHP](docs/images/OpenNHP_ACDemo0.png)

Le scan des ports du serveur retournera également une erreur de TYPE OUT.

![Démo OpenNHP](docs/images/OpenNHP_ScanDemo.png)

### 2) Après authentification, le serveur protégé devient accessible

OpenNHP supporte une variété de méthodes d'authentification, telles que OAuth, SAML, QR codes, et plus encore. Pour cette démonstration, nous utilisons un service d'authentification basé sur un nom d'utilisateur/mot de passe simple à l'adresse https://demologin.opennhp.org.

![Démo OpenNHP](docs/images/OpenNHP_DemoLogin.png)

Une fois que vous cliquez sur le bouton "Login", l'authentification est réussie, et vous êtes redirigé vers le serveur protégé. Le serveur devient alors *"visible"* et accessible sur votre appareil.

![Démo OpenNHP](docs/images/OpenNHP_ACDemo1.png)

---

## Vision : Faire d'Internet un espace de confiance

L'ouverture des protocoles TCP/IP a stimulé la croissance des applications Internet, mais a aussi exposé des vulnérabilités, permettant aux acteurs malveillants d'accéder de manière non autorisée à toute adresse IP exposée. Bien que le [modèle réseau OSI](https://fr.wikipedia.org/wiki/Mod%C3%A8le_OSI) définisse la *couche 5 (couche session)* pour la gestion des connexions, peu de solutions efficaces ont été mises en place à cet égard.

**NHP**, ou **"Protocole de Masquage de l'Infrastructure Réseau"**, est un protocole réseau de zéro confiance, basé sur la cryptographie, conçu pour fonctionner au *niveau de la couche session OSI*, idéal pour gérer la visibilité réseau et les connexions. L'objectif principal de NHP est de dissimuler les ressources protégées des entités non autorisées, accordant l'accès uniquement aux utilisateurs vérifiés et autorisés par une vérification continue, contribuant ainsi à un Internet plus digne de confiance.

![Internet de Confiance](docs/images/TrustworthyCyberspace.png)

---

## Solution : OpenNHP rétablit le contrôle de la visibilité réseau

**OpenNHP** est l'implémentation open source du protocole NHP. Il est basé sur la cryptographie et conçu avec des principes de sécurité en priorité, implémentant une véritable architecture de zéro confiance au *niveau de la couche session OSI*.

![OpenNHP en tant que couche 5 OSI](docs/images/OSI_OpenNHP.png)

OpenNHP s'appuie sur des recherches antérieures sur la technologie de masquage réseau, en utilisant des cadres et une architecture modernes de cryptographie pour garantir la sécurité et des performances élevées, surmontant ainsi les limitations des technologies précédentes.

| Protocole de Masquage de l'Infrastructure | 1ère Génération | 2ème Génération | 3ème Génération |
|:---|:---|:---|:---|
| **Technologie Clé** | [Port Knocking](https://fr.wikipedia.org/wiki/Port_knocking) | [Autorisation par Paquet Unique (SPA)](https://cloudsecurityalliance.org/artifacts/software-defined-perimeter-zero-trust-specification-v2) | Protocole de Masquage de l'Infrastructure Réseau (NHP) |
| **Authentification** | Séquences de ports | Secrets partagés | Cadre cryptographique moderne |
| **Architecture** | Pas de plan de contrôle | Pas de plan de contrôle | Plan de contrôle scalable |
| **Capacité** | Masquer les ports | Masquer les ports | Masquer les ports, IPs et domaines |
| **Contrôle d'Accès** | Niveau IP | Niveau Port | Niveau Application |
| **Projets Open Source** | [knock](https://github.com/jvinet/knock) *(C)* | [fwknop](https://github.com/mrash/fwknop) *(C++)* | [OpenNHP](https://github.com/OpenNHP/opennhp) *(Go)* |

> Il est crucial de choisir un langage **sûr pour la mémoire** comme *Go* pour le développement d'OpenNHP, comme le souligne le [rapport technique du gouvernement des États-Unis](https://www.whitehouse.gov/wp-content/uploads/2024/02/Final-ONCD-Technical-Report.pdf). Pour une comparaison détaillée entre **SPA et NHP**, référez-vous à la [section ci-dessous](#comparison-between-spa-and-nhp).

## Bénéfices en matière de sécurité

Puisqu'OpenNHP implémente les principes de zéro confiance au *niveau de la couche session OSI*, il offre des avantages significatifs :

- Réduit la surface d'attaque en cachant l'infrastructure
- Empêche la reconnaissance réseau non autorisée
- Atténue l'exploitation des vulnérabilités
- Empêche le phishing via DNS chiffré
- Protège contre les attaques DDoS
- Permet un contrôle d'accès granulaire
- Fournit un suivi des connexions basé sur l'identité
- Attribution des attaques

## Architecture

L'architecture d'OpenNHP s'inspire de la [norme d'architecture Zero Trust du NIST](https://www.nist.gov/publications/zero-trust-architecture). Elle suit une conception modulaire avec trois composants principaux : **NHP-Server**, **NHP-AC** et **NHP-Agent**, comme illustré dans le diagramme ci-dessous.

![Architecture OpenNHP](docs/images/OpenNHP_Arch.gif)

> Veuillez consulter la [documentation d'OpenNHP](https://docs.opennhp.org/) pour des informations détaillées sur l'architecture et le flux de travail.

## Cœur : Algorithmes Cryptographiques

La cryptographie est au cœur d'OpenNHP, fournissant une sécurité robuste, d'excellentes performances et une bonne évolutivité en utilisant des algorithmes cryptographiques de pointe. Voici les principaux algorithmes et cadres cryptographiques employés par OpenNHP :

- **[Cryptographie à Courbes Elliptiques (ECC)](https://fr.wikipedia.org/wiki/Cryptographie_sur_courbe_elliptique)** : Utilisée pour la cryptographie asymétrique efficace.

> Comparée à RSA, l'ECC offre une efficacité supérieure avec un chiffrement plus fort à des longueurs de clé plus courtes, améliorant la transmission réseau et les performances de calcul. Le tableau ci-dessous montre les différences de force de sécurité, de longueurs de clé et du ratio entre RSA et ECC, ainsi que leurs périodes de validité respectives.

| Force de Sécurité (bits) | Longueur de Clé DSA/RSA (bits) | Longueur de Clé ECC (bits) | Ratio : ECC vs DSA/RSA | Validité |
|:--------------------------:|:------------------------------:|:--------------------------:|:-----------------------:|:---------:|
| 80                         | 1024                           | 160-223                    | 1:6                     | Jusqu'en 2010 |
| 112                        | 2048                           | 224-255                    | 1:9                     | Jusqu'en 2030 |
| 128                        | 3072                           | 256-383                    | 1:12                    | Après 2031 |
| 192                        | 7680                           | 384-511                    | 1:20                    | |
| 256                        | 15360                          | 512+                       | 1:30                    | |

- **[Cadre de Protocole Noise](https://noiseprotocol.org/)** : Permet l'échange de clés sécurisé, le chiffrement/déchiffrement des messages, et l'authentification mutuelle.

> Le protocole Noise est basé sur l'[accord de clé Diffie-Hellman](https://fr.wikipedia.org/wiki/%C3%89change_de_cl%C3%A9_Diffie-Hellman) et offre des solutions cryptographiques modernes telles que l'authentification mutuelle et optionnelle, le masquage de l'identité, la sécurité persistante, et le chiffrement à tour de passezà-tour de zéro. Déjà prouvé pour sa sécurité et ses performances, il est utilisé par des applications populaires comme [WhatsApp](https://www.whatsapp.com/security/WhatsApp-Security-Whitepaper.pdf), [Slack](https://github.com/slackhq/nebula), et [WireGuard](https://www.wireguard.com/).

- **[Cryptographie basée sur l'Identité (IBC)](https://fr.wikipedia.org/wiki/Cryptographie_bas%C3%A9e_sur_l%27identit%C3%A9)** : Simplifie la distribution des clés à grande échelle.

> Une distribution efficace des clés est essentielle pour implémenter le Zéro Confiance. OpenNHP prend en charge à la fois PKI et IBC. Alors que PKI est utilisée depuis des décennies, elle dépend de Certificats d'Autorité centralisés (CA) pour la vérification de l'identité et la gestion des clés, ce qui peut être long et coûteux. En revanche, l'IBC permet une approche décentralisée et autonome de la vérification de l'identité et de la gestion des clés, la rendant plus rentable pour l'environnement Zero Trust d'OpenNHP, où des milliards d'appareils ou de serveurs peuvent avoir besoin de protection et d'intégration en temps réel.

- **[Cryptographie à Clé Publique sans Certificat (CL-PKC)](https://fr.wikipedia.org/wiki/Cryptographie_sans_certificat)** : Algorithme IBC recommandé

> CL-PKC est un schéma qui améliore la sécurité en évitant la garde des clés et en répondant aux limites de la cryptographie basée sur l'identité (IBC). Dans la plupart des systèmes IBC, la clé privée d'un utilisateur est générée par un Centre de Génération de Clés (KGC), ce qui introduit des risques importants. Un KGC compromis peut entraîner l'exposition des clés privées de tous les utilisateurs, nécessitant une confiance totale dans le KGC. CL-PKC atténue ce problème en divisant le processus de génération de clés, de sorte que le KGC n'a connaissance que d'une clé privée partielle. En conséquence, CL-PKC combine les forces du PKI et de l'IBC, offrant une sécurité renforcée sans les inconvénients de la gestion centralisée des clés.

Pour en savoir plus :

> Veuillez consulter la [documentation OpenNHP](https://docs.opennhp.org/cryptography/) pour une explication détaillée des algorithmes cryptographiques utilisés dans OpenNHP.

## Principales Fonctionnalités

- Atténue l'exploitation des vulnérabilités en appliquant par défaut des règles "deny-all"
- Empêche les attaques de phishing via la résolution DNS chiffrée
- Protège contre les attaques DDoS en cachant l'infrastructure
- Permet l'attribution des attaques via des connexions basées sur l'identité
- Contrôle d'accès par défaut pour toutes les ressources protégées
- Authentification basée sur l'identité et les appareils avant l'accès au réseau
- Résolution DNS chiffrée pour empêcher le piratage DNS
- Infrastructure distribuée pour atténuer les attaques DDoS
- Architecture évolutive avec des composants découplés
- Intégration avec les systèmes existants de gestion des identités et des accès
- Prend en charge divers modèles de déploiement (client-à-passerelle, client-à-serveur, etc.)
- Sécurité cryptographique avec des algorithmes modernes (ECC, Noise Protocol, IBC)

<details>
<summary>Cliquez pour développer les détails des fonctionnalités</summary>

- **Contrôle d'accès par défaut** : Toutes les ressources sont cachées par défaut, ne devenant accessibles qu'après authentification et autorisation.
- **Authentification basée sur l'identité et les appareils** : Garantit que seuls les utilisateurs connus sur des appareils approuvés peuvent accéder.
- **Résolution DNS chiffrée** : Empêche le piratage DNS et les attaques de phishing associées.
- **Atténuation des DDoS** : Conception d'infrastructure distribuée aide à protéger contre les attaques par DDoS.
- **Architecture évolutive** : Les composants découplés permettent un déploiement et une évolution flexibles.
- **Intégration IAM** : Fonctionne avec vos systèmes de gestion des identités et des accès.
- **Déploiement flexible** : Prend en charge divers modèles, y compris client-à-passerelle, client-à-serveur, et plus encore.
- **Cryptographie forte** : Utilise des algorithmes modernes comme ECC, Noise Protocol, et IBC pour une sécurité robuste.
</details>

## Déploiement

OpenNHP prend en charge plusieurs modèles de déploiement pour répondre à différents cas d'utilisation :

- Client-à-Passerelle : Sécurise l'accès à plusieurs serveurs derrière une passerelle
- Client-à-Serveur : Sécurise directement des serveurs/applications individuels
- Serveur-à-Serveur : Sécurise la communication entre les services backend
- Passerelle-à-Passerelle : Sécurise les connexions site-à-site

> Veuillez consulter la [documentation OpenNHP](https://docs.opennhp.org/deploy/) pour des instructions de déploiement détaillées.

## Comparaison entre SPA et NHP
Le protocole d'Autorisation par Paquet Unique (SPA) est inclus dans la [spécification du Périmètre Défini par Logiciel (SDP)](https://cloudsecurityalliance.org/artifacts/software-defined-perimeter-zero-trust-specification-v2) publiée par l'[Alliance pour la Sécurité Cloud (CSA)](https://cloudsecurityalliance.org/). NHP améliore la sécurité, la fiabilité, la scalabilité et l'extensibilité grâce à un cadre et une architecture de cryptographie modernes, comme démontré dans l'article de recherche [AHAC](https://www.mdpi.com/2076-3417/14/13/5593).

| - | SPA | NHP | Avantages de NHP |
|:---|:---|:---|:---|
| **Architecture** | Le déchiffrement du paquet SPA et le composant d'authentification de l'utilisateur/appareil sont couplés au composant de contrôle d'accès réseau dans le serveur SPA. | NHP-Server (le composant de déchiffrement de paquet et d'authentification utilisateur/appareil) et NHP-AC (le composant de contrôle d'accès) sont découplés. NHP-Server peut être déployé sur des hôtes distincts et prend en charge la mise à l'échelle horizontale. | <ul><li>Performance : le composant gourmand en ressources NHP-server est séparé du serveur protégé.</li><li>Scalabilité : NHP-server peut être déployé en mode distribué ou en cluster.</li><li>Sécurité : l'adresse IP du serveur protégé n'est pas visible par le client tant que l'authentification n'a pas réussi.</li></ul>|
| **Communication** | Simple direction | Bidirectionnelle | Meilleure fiabilité avec la notification d'état du contrôle d'accès |
| **Cadre cryptographique** | Secrets partagés | PKI ou IBC, Cadre Noise | <ul><li>Sécurité : mécanisme éprouvé d'échange de clés pour atténuer les menaces MITM</li><li>Coût faible : distribution efficace des clés pour le modèle de zéro confiance</li><li>Performance : chiffrement/déchiffrement haute performance</li></ul>|
| **Capacité de Masquage de l'Infrastructure Réseau** | Uniquement les ports de serveur | Domaines, IP et ports | Plus puissant contre diverses attaques (e.g., vulnérabilités, piratage DNS, et attaques DDoS) |
| **Extensibilité** | Aucune, uniquement pour SDP | Tout usage | Prise en charge de tout scénario nécessitant un obscurcissement de service |
| **Interopérabilité** | Non disponible | Personnalisable | NHP peut s'intégrer de manière transparente avec les protocoles existants (e.g., DNS, FIDO, etc.) |

## Contribuer

Nous accueillons avec plaisir les contributions à OpenNHP ! Veuillez consulter nos [lignes directrices de contribution](CONTRIBUTING.md) pour plus d'informations sur la manière de participer.

## Licence

OpenNHP est publié sous la [licence Apache 2.0](LICENSE).

## Contact

- Site Web du Projet : [https://github.com/OpenNHP/opennhp](https://github.com/OpenNHP/opennhp)
- E-mail : [opennhp@gmail.com](mailto:opennhp@gmail.com)
- Discord : [Rejoignez notre Discord](https://discord.gg/CpyVmspx5x)

Pour plus de documentation détaillée, veuillez visiter notre [Documentation Officielle](https://opennhp.org).

## Références

- [Spécification du Périmètre Défini par Logiciel (SDP) v2.0](https://cloudsecurityalliance.org/artifacts/software-defined-perimeter-zero-trust-specification-v2). Jason Garbis, Juanita Koilpillai, Junaid lslam, Bob Flores, Daniel Bailey, Benfeng Chen, Eitan Bremler, Michael Roza, Ahmed Refaey Hussein. [*Cloud Security Alliance (CSA)*](https://cloudsecurityalliance.org/). Mar 2022.
- [AHAC : Cadre Avancé de Contrôle d'Accès Caché au Réseau](https://www.mdpi.com/2076-3417/14/13/5593). Mudi Xu, Benfeng Chen, Zhizhong Tan, Shan Chen, Lei Wang, Yan Liu, Tai Io San, Sou Wang Fong, Wenyong Wang, et Jing Feng. *Journal des Sciences Appliquées*. Juin 2024.
- [STALE : Un schéma d'authentification transfrontalière évolutif et sécurisé tirant parti du courrier électronique et de l'échange de clés ECDH](https://www.mdpi.com/2079-9292/14/12/2399) Jiexin Zheng, Mudi Xu, Jianqing Li, Benfeng Chen, Zhizhong Tan, Anyu Wang, Shuo Zhang, Yan Liu, Kevin Qi Zhang, Lirong Zheng, et Wenyong Wang. *électronique*. Juin 2025.
- [DRL-AMIR : Planification intelligente des flux pour les réseaux de confiance zéro définis par logiciel](https://www.techscience.com/cmc/v84n2/62920) WenLong Ke, Zilong Li, Peiyu Chen, Benfeng Chen, Jinglin Lv, Qiang Wang, Ziyi Jia et Shigen Shen. *CMC*. Juillet 2025.
[méthode de contrôle du trafic réseau de NHP basée sur l’apprentissage par renforcement profond](https://www.nature.com/articles/s41598-025-31556-3). Qinglin Huang, Zhizhong Tan, Qiang Wang, Ziyi Jia et Benfeng Chen. *rapports scientifiques par Nature* décembre 2025.
- Noise Protocol Framework. https://noiseprotocol.org/
- Projet de Cadre de Gestion des Vulnérabilités. https://phoenix.security/web-vuln-management/

---

🌟 Merci pour votre intérêt dans OpenNHP ! Nous attendons vos contributions et vos commentaires avec impatience.

