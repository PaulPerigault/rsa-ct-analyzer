---
name: encre-securite
description: Revue de sécurité (CSP, en-têtes, attribut `style`, dépendance, workflow GitHub, publication npm, entrée utilisateur, hook, agent, WebAssembly, skill tiers). À charger pour relire ou écrire ces sujets ; « revue sécurité », « chaîne d'approvisionnement ».
---

# Encre : revue de sécurité

Issue : [#178](https://github.com/PaulPerigault/design-system/issues/178). Sources du dépôt :
`SECURITY.md`, ADR-0004, ADR-0005, ADR-0006, `apps/docs/scripts/headers/policy.mjs`, D-005, D-024,
D-030, D-056, [#176](https://github.com/PaulPerigault/design-system/issues/176). Sources externes :
section « Sources ». Ce skill est de la méthode, pas une preuve : « prouvé » seulement si une source
l'établit ; sinon « convention » ou « préférence ». Une faille ne se signale jamais dans une issue.

## Comment l'utiliser

- Phrase de déclenchement : « revue sécurité de `<fichier, PR ou paquet>` » (ou `/encre-securite`).
- Lire le code avant de juger ; citer `fichier:ligne` pour chaque remarque.
- Sortie : trois sections (Critique, À améliorer, Ce qui tient), gravité en tête de point.
- Il ne remplace ni CodeQL, ni `npm audit`, ni `secretlint` : il dit quoi regarder en plus.

## CSP sans `unsafe-inline` (ADR-0004)

- Les consommateurs ont une CSP stricte (`style-src` par hachages, sans `unsafe-inline`) : un attribut
  `style` en ligne est ignoré : règle et variantes dans `AGENTS.md`.
- Politique du site : `apps/docs/scripts/headers/policy.mjs`, `default-src 'none'`, `script-src 'self'
'wasm-unsafe-eval'` plus hachages, `style-src 'self'` plus hachages, `object-src 'none'`,
  `base-uri 'none'`, `frame-ancestors 'none'`. Ni `unsafe-inline`, ni `unsafe-eval`.
- CSS servi en fichiers (D-024, `inlineStylesheets: 'never'`) : pas de hachage de style par page.
- Un contenu en ligne futur est couvert par `hashes.mjs` (SHA-256) ; tout changement d'un octet
  d'un script change son hachage (MDN, OWASP : paraphrase, niveau norme et convention).
- Relire : un `<style>`, un `style=`, un `on*=`, un `javascript:`, un `eval`, une ressource tierce.
- Toute directive élargie (une source ajoutée) demande une décision écrite (D-xxx).

## Chaîne d'approvisionnement

- Publication : seulement par le workflow `release` (trusted publishing OIDC, provenance npm, aucun
  `NPM_TOKEN`). Jamais `npm publish` à la main ; `npm pack --dry-run` pour vérifier le contenu.
- Consommateur : `npm audit signatures` vérifie les signatures du registre et les attestations de
  provenance (documentation npm). En CI : `supply-chain.yml` et `release.yml`.
- Installation : `--ignore-scripts`, liste blanche (`tools/checks/check-install-scripts.mjs`), registre
  unique npmjs (`.npmrc`, `check:lockfile` : aucune dépendance `file:` ni `link:`).
- Versions : `package.json` accepte des plages (`^`), le `package-lock.json` fixe les versions
  installées ; les actions GitHub sont épinglées par SHA (`SECURITY.md`). « Versions exactes » dans
  `package.json` n'est pas une règle écrite du dépôt : proposition, à valider par Paul.
- Mises à jour : Dependabot avec délai de 3 jours (7 pour une majeure) ; majeures relues à la main.
- SBOM CycloneDX par paquet (`tools/release/sbom.mjs`), artefact de la release et de `supply-chain`.
- SLSA : la provenance npm ressemble au niveau 2 de la piste Build (plateforme hébergée qui signe
  la provenance) ; le dépôt ne revendique aucun niveau SLSA. Ne jamais en écrire un sans mesure.
- OpenSSF Scorecard : `scorecard.yml`, note publiée pour le badge. Une note n'est pas un audit.
- Relire une PR de dépendance : le mainteneur, la provenance, les scripts d'installation, le poids.

## Validation aux frontières (ADR-0006)

- Tout ce qui entre (JSON, arguments de CLI, fichier, réponse) est validé à la frontière en un type
  du domaine (DTO), avec une erreur typée ; aucune erreur anonyme ne franchit une frontière.
- Liste blanche, pas liste de refus : définir ce qui est accepté, refuser le reste (OWASP).
- Contrôle de forme (type, longueur, format) puis de sens (plage, cohérence entre champs).
- Relire : un `JSON.parse` sans schéma, un `as`, un `any`, un `throw` de chaîne, un chemin de
  fichier venu de l'entrée sans normalisation.

## Injection CSS (D-056)

- `--accent-hue` : nombre fini entre 0 et 360. `--font-stack` : 256 caractères au plus, refus de `;`,
  accolades, `url(`, `@import`, retours à la ligne, commentaires, guillemets non appariés
  (`stack-unsafe`, `parseBuildInput`).
- Règle OWASP : une donnée non fiable n'entre que dans une valeur de propriété, jamais dans un
  sélecteur ni une déclaration entière ; la propriété est fixe, la valeur vient d'une liste admise.
- Relire : une chaîne d'entrée concaténée dans du CSS, une propriété choisie par l'appelant.

## Secrets

- Aucun secret dans le dépôt : `secretlint` (`npm run check:secrets`, dans `guard`). `.env*` ignorés.
- Aucun secret transmis à Kiro ; aucune variable sensible lue par un agent (`AGENTS.md`).
- Publication sans jeton stocké (OIDC). Une fuite : révoquer d'abord, puis signaler en privé.

## Hooks et agents (#176)

- Liste blanche plutôt que liste de refus : une liste de refus laisse passer toute commande non
  prévue (autre binaire réseau, interpréteur, redirection). Cible de #176 : `kiro-hook.ts` n'accepte
  que des commandes nommées. D-005 : Kiro sans `--trust-all-tools`, liste fermée.
- État au 2026-10-30 : `tools/agents/lib/forbidden.ts` est encore une liste de refus (#176 ouverte).
- Un agent n'écrit que les fichiers de `files:`, jamais de `git commit`, `push`, `npm publish`.
- Relire un hook : que se passe-t-il pour une commande inconnue ? La réponse doit être « refusée ».

## WebAssembly

- `script-src` porte `'wasm-unsafe-eval'` seulement, jamais `'unsafe-eval'` : le premier n'autorise que
  WebAssembly, le second rouvre aussi `eval()` et `Function()` (MDN).
- Le `.wasm` est compilé en CI et publié dans le paquet npm (D-030) : même provenance qu'un paquet,
  mêmes contrôles (`npm audit signatures`, SBOM). Préférence à valider : aucun `.wasm` commité sans
  sa source. Le seul usage actuel est la recherche Pagefind (commentaire de `policy.mjs`).
- Poids du `.wasm` budgété (D-030). Aucune ressource `.wasm` chargée depuis un hôte tiers.

## Rust (étape 7, D-030)

- Aucun `panic` ni `unwrap` hors tests (D-030). Erreurs typées.
- Dépendances : `cargo audit` lit `Cargo.lock` contre la base RustSec ; `cargo deny` vérifie avis de
  sécurité, bannissements, licences et sources des crates (ADR-0005 prévoit les deux en CI).
- `unsafe` : `#![forbid(unsafe_code)]` par défaut (`forbid` ne se contourne pas par `allow` local,
  contrairement à `deny`). Une exception demande ⚖️ de Paul : le bloc `unsafe` est justifié par un
  commentaire qui énonce le contrat. Préférence du projet, non écrite dans une ADR : à valider.

## Skills tiers : provenance (D-116)

Un skill est du texte exécuté avec les droits de Paul (shell, écriture, `gh`) : une chaîne
d'approvisionnement. Rapport : `docs/design/veille-skills-2026-11.md`. Préférence de Paul, non norme.

1. Jamais `npx`, `npm -g`, `curl`, `/plugin marketplace add` ni installeur tiers : ils exécutent du code
   et posent des hooks. Clone en lecture seule hors du dépôt (quarantaine du bloc-notes), SHA épinglé,
   chaque fichier relu. Aucun script de candidat n'est lancé.
2. Grille de relecture : commande shell, appel réseau, `curl`, `wget`, `eval`, `base64`, binaire à
   télécharger, hook, `allowed-tools` large, télémétrie, script joint, « ignore tes règles », « ne
   demande pas confirmation », « mode obligatoire », écriture hors du dossier du skill, fichier de
   notes créé. Un signal rouge non neutralisable : refusé.
3. Licence lue dans le dépôt. MIT : texte et copyright du titulaire. Apache-2.0 : licence, `NOTICE`,
   modifications mentionnées. CC BY 4.0 : auteur, lien, modifications. Absente, ambiguë ou sans
   titulaire : refusé.
4. Jamais copié tel quel : réécrit au format `encre-*` (150 lignes au plus, description de 300
   caractères au plus qui dit QUAND charger), « Surcharges Encre » en tête (une PR à la fois, aucun
   sous-agent, design avant code, ni ombre, ni dégradé, ni flou décoratif, zéro dépendance inutile,
   ligne finale). Chaque skill porte `PROVENANCE.md` : URL, SHA, licence, date, modifications.
5. Plafond : 12 skills de projet, descriptions toujours chargées. Un skill n'entre que si un autre
   sort ou si son apport est supérieur. Dépasser : ⚖️ dans `tmp/questions-paul.md`.
6. Installation dans `.claude/skills/` du projet, jamais au niveau utilisateur ; index d'`AGENTS.md`
   (lire seulement ceux de la tâche). Aucun hook, script ni paquet ajouté pour un skill.

## Sources

Consultées le 2026-10-30. Niveaux : norme, étude, convention, préférence.

| Source                                                                    | URL                                                                                                    | Niveau       |
| ------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------ | ------------ |
| MDN, Content Security Policy                                              | https://developer.mozilla.org/en-US/docs/Web/HTTP/Guides/CSP                                           | convention   |
| MDN, `script-src` (`'wasm-unsafe-eval'`)                                  | https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Content-Security-Policy/script-src | convention   |
| OWASP, CSP Cheat Sheet                                                    | https://cheatsheetseries.owasp.org/cheatsheets/Content_Security_Policy_Cheat_Sheet.html                | convention   |
| OWASP, Input Validation Cheat Sheet                                       | https://cheatsheetseries.owasp.org/cheatsheets/Input_Validation_Cheat_Sheet.html                       | convention   |
| OWASP, XSS Prevention Cheat Sheet (contexte CSS)                          | https://cheatsheetseries.owasp.org/cheatsheets/Cross_Site_Scripting_Prevention_Cheat_Sheet.html        | convention   |
| OWASP ASVS 5.0.0 (30 mai 2025)                                            | https://asvs.owasp.org/                                                                                | norme        |
| npm, provenance                                                           | https://docs.npmjs.com/generating-provenance-statements                                                | convention   |
| npm, `npm audit` (signatures)                                             | https://docs.npmjs.com/cli/v11/commands/npm-audit                                                      | convention   |
| SLSA v1.0, niveaux de la piste Build                                      | https://slsa.dev/spec/v1.0/levels                                                                      | norme        |
| OpenSSF Scorecard                                                         | https://github.com/ossf/scorecard                                                                      | convention   |
| RustSec et `cargo audit`                                                  | https://rustsec.org/                                                                                   | convention   |
| cargo-deny, les quatre contrôles                                          | https://embarkstudios.github.io/cargo-deny/checks/index.html                                           | convention   |
| Rust, lint `unsafe_code`                                                  | https://doc.rust-lang.org/rustc/lints/listing/allowed-by-default.html#unsafe-code                      | norme        |
| Rustonomicon, `unsafe`                                                    | https://doc.rust-lang.org/nomicon/meet-safe-and-unsafe.html                                            | convention   |
| ANSSI, guide « maîtriser les standards de sécurité côté navigateur » v2.0 | [À VÉRIFIER] l'adresse issue d'une recherche renvoie 404 ; le document n'a pas été lu                  | [À VÉRIFIER] |

« Norme » : spécification ou standard publié. « Convention » : guide de praticiens reconnu, non
normatif. ASVS est un standard communautaire de l'OWASP. Tout est paraphrasé ; aucune étude citée.
