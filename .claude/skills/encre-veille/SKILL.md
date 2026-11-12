---
name: encre-veille
description: Chercher, évaluer et installer un skill tiers sous la règle de provenance (D-116) et le plafond de 12 ; manque constaté, correction répétée de Paul, nouvelle technologie, revue mensuelle. « veille de skills », « un skill pour ça ? ».
---

# Encre : veille de skills

Issue : [#222](https://github.com/PaulPerigault/design-system/issues/222). Sources du dépôt :
`encre-securite` (règle de provenance, D-116), `docs/design/skills-registre.md`,
`docs/design/veille-skills-2026-11.md`, `tools/skills/check-upstream.mjs`,
`.github/workflows/skills-amont.yml`. La méthode de suivi des corrections est adaptée de « Task
Observer » (voir « Attribution »). En cas d'écart, `AGENTS.md` et `encre-securite` gagnent.

## Surcharges Encre

- Aucune installation automatique, aucune tâche récurrente qui écrit, aucune issue ni PR créée par la
  veille : la veille propose, Paul décide. Propositions dans `tmp/questions-paul.md`.
- Ni hook, ni dossier d'observations, ni fichier de notes dans le dépôt ; aucun script de candidat
  exécuté ; aucun sous-agent ; une PR à la fois, une par skill ; écriture seulement dans
  `.claude/skills/`, `docs/`, `AGENTS.md` et `.github/`.
- Pas de phrase d'accueil imposée : on répond à la demande.

## Quand chercher

1. **Manque constaté** : une tâche n'a aucun skill adapté dans l'index d'`AGENTS.md`.
2. **Correction répétée** : Paul corrige deux fois la même chose.
3. **Nouvelle technologie** : un nouveau paquet, outil ou format entre dans le projet.
4. **Revue mensuelle** : lire le résumé du job « Skills amont » ; pour chaque ligne « changé »,
   ouvrir le lien de comparaison, relire le diff avec la grille, décider : mettre à jour ou ignorer.
   Le SHA de `HEAD` bouge à chaque commit amont : un « changé » demande un coup d'œil, pas une réécriture.

Pas de déclencheur : l'envie d'un skill « au cas où ». Une nouvelle liste donnée par Paul est ajoutée
au registre puis évaluée comme le reste.

## Comment chercher

1. **Registre d'abord** (`skills-registre.md`) : candidat déjà vu ? Ne pas le réévaluer sans raison
   (licence changée, nouvelle version majeure, correction du motif de refus).
2. **Sources** : documentation officielle du fournisseur, dépôts d'organisations identifiées, dépôts
   très actifs à licence claire. Outils : recherche web (mode standard d'abord), `gh api repos/<dépôt>`
   en lecture seule, clone en quarantaine hors du dépôt pour lire les fichiers.
3. **Critères d'écart, avant de lire** : licence SPDX claire avec titulaire ; dernière activité sous six
   mois ; mainteneur identifiable ; ni hook, ni binaire, ni télémétrie, ni installeur ; un manque réel
   couvert. Les étoiles mesurent la popularité, jamais l'acceptation. Abandonné ou sans licence :
   ignoré, mais noté au registre.
4. **Relecture** par la grille de `encre-securite` ; verdict **retenu**, **proposé**, **refusé** ou
   **non lu**, avec SHA et date dans le registre. Une URL non ouvrable : `[À VÉRIFIER]`.

## Grille et règle (rappel ; la source est `encre-securite`)

Signaux rouges : shell, réseau, `curl`, `wget`, `eval`, `base64`, binaire à télécharger, hook,
`allowed-tools` large, télémétrie, script joint, consigne de contournement, écriture hors du skill.
Jamais `npx`, `npm -g`, `curl`, `/plugin marketplace add`. Licence lue et copiée. Skill réécrit, jamais
copié, avec « Surcharges Encre » et `PROVENANCE.md`. Installation dans `.claude/skills/` du projet.

## Plafond de 12 et remplacement

- Les descriptions des skills sont toujours chargées : 12 skills de projet au plus, description de
  300 caractères au plus, 150 lignes par `SKILL.md`. Compter avant toute proposition.
- Un skill n'entre que si un autre sort, ou si son apport est supérieur. Comparer sur quatre points :
  **manque couvert** (aucun skill ni ADR n'y répond), **fréquence** des tâches concernées, **coût de
  contexte**, **non-redondance** (ne double ni un skill ni un ADR).
- Au plafond : proposer lequel sort, avec la comparaison, ou une fusion qui tient en 150 lignes.
  Paul tranche (⚖️) ; sans accord, rien n'est installé.

## Corrections de Paul (méthode adaptée de Task Observer)

Signaux à repérer pendant une tâche :

- **Nouveau skill** : un flux qui revient, une méthode que Paul explique et qu'aucun skill ne porte.
- **Amélioration** : une règle écrite est violée (renforcer par un contrôle, pas par des majuscules) ;
  une correction révèle une règle manquante ; le réel contredit le skill.
- **Simplification** : une section jamais utile, une règle née d'une seule observation, deux règles
  qui se contredisent.

Test de généralité : la correction vaut-elle pour d'autres tâches et d'autres fichiers ? Sinon c'est une
préférence ponctuelle, on n'écrit rien. À la deuxième violation d'une règle, ajouter une barrière
(contrôle `check:*`, test), non un rappel plus long.

La proposition va dans `tmp/questions-paul.md` : les deux corrections citées, le skill concerné, le
changement proposé, le verdict recommandé. Jamais appliquée sans accord de Paul.

## Format de rapport

1. **Déclencheur** et date. 2. **Recherche** : requêtes, sources, date de consultation. 3. **Candidats** :
   tableau (URL, SHA, licence, étoiles, dernière activité, constats de la grille, verdict, raison).
2. **Proposition** : installer, remplacer qui, ignorer. 5. **Question pour Paul** avec recommandation.
   Niveaux de preuve : lecture directe, mesure, présomption. Une présomption n'est jamais écrite « prouvé ».

## Après un accord

Branche, réécriture au format `encre-*`, `PROVENANCE.md`, lignes du registre (« Suivi amont » et
« Candidats vus »), index d'`AGENTS.md`, une ligne D-xxx, journal, une PR (`encre-processus`).

## Attribution

La méthode de suivi des corrections est adaptée de « Task Observer » (Eoghan Henn, https://rebelytics.com),
dépôt https://github.com/rebelytics/one-skill-to-rule-them-all, SHA
`ba696b2c2673be959dcd4a1dbdbc92085df8b41a`, licence CC BY 4.0
(https://creativecommons.org/licenses/by/4.0/). Modifié : traduit, réduit, privé du journal
d'observations, des scripts et du balayage au démarrage. Voir `PROVENANCE.md`.
