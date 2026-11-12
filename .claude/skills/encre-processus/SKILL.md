---
name: encre-processus
description: Processus de livraison (issue, branche, commit, PR, fusion en squash, journal, décisions, fichiers générés, règles de saut). À charger avant une branche, un commit, un titre ou corps de PR, une fusion, `docs.json` ou `adr.json` ; « comment je livre ».
---

# Encre : processus de livraison

Issue : [#178](https://github.com/PaulPerigault/design-system/issues/178). Ce skill ne crée aucune
règle : chaque point renvoie à sa source dans le dépôt. En cas d'écart, la source gagne.
Sources : `AGENTS.md` (Git, GitHub, Fichiers générés, Décisions), `commitlint.config.mjs`,
`.github/workflows/ci.yml`, `.github/PULL_REQUEST_TEMPLATE.md`, `docs/process/templates.md`,
`.github/rulesets/`, `.husky/`, et `tmp/etat-autonome.md` (état local du run, jamais commité).

## Comment l'utiliser

- Phrase de déclenchement : « processus Encre » (ou `/encre-processus`).
- Lire la section qui correspond à l'étape en cours, pas tout le fichier.
- Un doute sur une règle : ouvrir la source citée, ne pas se fier à ce résumé.
- Ce skill est de la méthode, pas une preuve : il n'ajoute aucune exigence à la CI.

## Chaîne d'une étape

- Une étape = une issue = une PR (D-035, `AGENTS.md`). Chaîne : issue, branche
  `<type>/<n>-<slug>`, commits, PR, Project.
- Issue créée par un formulaire de `.github/ISSUE_TEMPLATE/`, complète : assignee, labels, projet,
  milestone, description. Les issues vierges sont refusées. Une faille ne se signale jamais dans
  une issue (`SECURITY.md`).
- Labels : seulement ceux de `.github/labels.json` (`node tools/github/sync-labels.mjs`, `--apply`
  pour agir). Jamais de label créé à la main.
- Pas de commentaire dans les issues ni dans les PR : l'information va dans la description de la PR
  ou dans le journal (`AGENTS.md`).
- `main` seule. Le ruleset exige une PR, la fusion en squash, un historique linéaire et le contrôle
  `ci` vert (`.github/rulesets/`). Aucun push direct, aucun push forcé.

## Design avant code (D-101)

- Une tâche qui crée ou déplace des modules commence par l'arborescence cible, écrite dans le
  journal du jour : chaque fichier, sa responsabilité en une phrase, ses dépendances autorisées
  (ADR-0006 : `infrastructure → application → domain` ; un crate `-wasm` dépend du crate métier,
  jamais l'inverse). Ensuite seulement, le code.
- Chercher une brique existante avant d'écrire (le paquet, puis les paquets en amont).
- Extraire seulement quand une deuxième occurrence existe, jamais par anticipation.
- Jamais de module « utils » ou « helpers » fourre-tout : un module porte le nom de sa responsabilité.

## Mots-clés de fermeture

- Une issue qui demande plusieurs PR : « Closes #n » seulement dans la **dernière**, une seule
  fois, en **première ligne** du corps. Les autres PR portent « Refs #n » (`tmp/etat-autonome.md`
  §1 ; le gabarit de PR met `Closes #` en tête de l'« Issue liée »).
- Aucun mot-clé (« closes », « fixes », « resolves ») dans la prose : GitHub le lit et fermerait
  l'issue trop tôt. Écrire « voir #n » ou « suite de #n ».

## Commits et titre de PR

- Conventional Commits. Types admis : `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`,
  `ci`, `perf`, `revert`. Pas de `build` (`commitlint.config.mjs`).
- Sujet en minuscules, 72 caractères au plus (`subject-case`, `subject-max-length`).
- Le suffixe « (#n) » compte dans les 72 caractères du sujet, mesuré par commitlint après
  `type(scope): ` (préfixe non compté). Le titre saisi n'en porte pas ; le suffixe s'écrit dans
  `--subject` à la fusion. Jobs `commitlint` et `squash-message`. Recommandé, non contrôlé : ligne
  entière sous 100 caractères (lisibilité de `git log`).
- Aucune signature d'outil : ni `Co-authored-by`, ni « Generated with » (règle `no-tool-signature`,
  D-004, sans exemption).
- Hooks locaux : `commit-msg` (commitlint) et `pre-commit` (lint-staged). Jamais `--no-verify` ni
  `HUSKY=0`. Il n'y a plus de hook `pre-push` : `npm run verify` se lance à la main avant l'envoi.

## Corps de PR

- Gabarit : `.github/PULL_REQUEST_TEMPLATE.md` (quoi et pourquoi, issue liée, type, vérifications,
  impact visuel ou accessibilité, notes pour la fusion).
- 90 colonnes au plus par ligne, mesurées en caractères et non en octets (les accents comptent un).
- Le titre et le corps passent par commitlint (job `commitlint` de la CI) : ils deviennent le
  message du commit de squash.

## Fusion

- Squash, avec le titre et le corps de la PR, jamais la liste des commits :
  `gh pr merge <n> --squash --delete-branch --subject "<titre> (#<n>)" --body-file <fichier>`.
- Sans `--subject` et `--body-file`, GitHub reprend les commits et `squash-message` échoue sur
  `main` (`AGENTS.md`).
- Une seule PR ouverte à la fois, fusionnée avant la suivante. Pas de fusion différée avec
  `--auto` : la consigne du run est de fusionner dès que `ci` est vert.
- La PR de release-please n'est jamais fusionnée par un agent : elle revient au propriétaire.

## Fichiers générés

- `docs.json`, `components.json`, `adr.json` (`packages/mcp/data/`) : commités seulement quand
  `check:data` ou `check:mcp-data` l'exige après un changement de source **réel**. Restaurés
  (`git checkout -- <fichier>`) quand ils diffèrent sans cause (dérive Windows) (`AGENTS.md`).
- `docs.json` lit `apps/docs/dist` : reconstruire la doc (`npm -w @paulperigault/docs run build`)
  **avant** `npm -w @paulperigault/mcp run build:data` (`packages/mcp/scripts/build-data.ts`,
  `SOURCES.dist`). `adr.json` lit `docs/adr/` et `docs/decisions-log.md` : toute nouvelle ligne D-xxx
  change `adr.json`.
- `npm run registry` régénère `registry.json` et les `meta.tokens` ; `npm run check:registry` le vérifie.
- Références Visuel : régénérées seulement pour les pages modifiées (workflow `visual.yml`).
- `package-lock.json` : modifié seulement si la tâche l'exige (`npm run check:lockfile`).

## Fichiers locaux

- `tmp/` n'est jamais commité : il est exclu par `.git/info/exclude` (local, absent de
  `.gitignore`). Sur un autre clone il serait visible : ajouter les fichiers par leur nom.
- `format:check` échoue localement sur `tmp/` et réussit en CI (`docs/journal/2026-10-30.md`).
- `tmp/questions-paul.md` : blocages à trancher par Paul. `tmp/etat-autonome.md` : état du run.

## Règles de saut (run autonome)

Source : `tmp/etat-autonome.md` §1. Une tâche est **bloquée** si :

- une décision revient à Paul (⚖️) ou une information manque (❓) ;
- elle échoue après deux essais (❌) ;
- la CI reste rouge après deux corrections ;
- une écriture GitHub est refusée, ou une étape est manuelle.

Alors : noter le blocage dans `tmp/questions-paul.md` (options et recommandation), laisser le dépôt
propre, passer à la tâche indépendante suivante, ou s'arrêter s'il n'y en a pas.

Jamais, quelle que soit la tâche : `npm publish`, `apply` OpenTofu, secret, push forcé, PR
release-please, et les issues exclues par l'état courant.

## Après chaque fusion

1. `docs/journal/AAAA-MM-JJ.md` : une entrée (branche, ce qui change, non-régression, suite).
2. `docs/decisions-log.md` : une ligne D-xxx (date, décision, alternative écartée, retour arrière),
   numéro suivant le dernier ; régénérer `adr.json` (voir « Fichiers générés »).
3. `tmp/etat-autonome.md` : dernière tâche finie, dernier D-xxx, PR fusionnées, suite.
4. Revenir sur `main`, `git pull`, supprimer la branche locale, `git status` propre.

Le journal du jour non commité part avec la première PR suivante. `docs/run-report.md` n'est mis à
jour qu'en fin de jalon (D-080) ; l'historique par PR est dans `docs/history/`.

Lecture initiale d'une session : `tmp/etat-autonome.md`, `AGENTS.md`, les skills de l'index, l'issue.

## Vérifier avant d'envoyer

```
npm run verify                          # guards, lint, format, tests, build, mcp, pack
npm run check:size                      # taille de fichier (limite d'`AGENTS.md`)
```

- Corps de PR : mesurer chaque ligne avant `gh pr create` (90 colonnes, en caractères).
- Titre : mesurer le sujet (après `type(scope): `) avec « (#n) » : 72 caractères au plus.
- Les limites de taille sont celles d'`AGENTS.md` (on découpe, on ne les relève jamais). Ce skill
  respecte la sienne : 150 lignes au plus, fixées par #178.

## Corrections répétées, skill manquant (D-124)

Paul corrige deux fois la même chose, ou une tâche n'a aucun skill adapté : recherche ciblée (recherche
web, puis lecture des fichiers), candidats et verdict recommandé dans `tmp/questions-paul.md`,
installation seulement après la grille de `encre-securite` (méthode : `encre-veille`). Une nouvelle
liste de Paul entre au registre, puis est évaluée de même. Aucun texte tiers n'est copié ici.
