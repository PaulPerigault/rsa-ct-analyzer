# Provenance : encre-veille

Skill réécrit, jamais copié tel quel. Règle : `encre-securite`, « Skills tiers : provenance » (D-116).
Issue : [#222](https://github.com/PaulPerigault/design-system/issues/222). Consulté le 2026-11-02.

| Source                                                       | URL                                                      | SHA lu                                     | Licence   |
| ------------------------------------------------------------ | -------------------------------------------------------- | ------------------------------------------ | --------- |
| « Task Observer » (signaux, test de généralité), Eoghan Henn | https://github.com/rebelytics/one-skill-to-rule-them-all | `ba696b2c2673be959dcd4a1dbdbc92085df8b41a` | CC BY 4.0 |

## Attribution (CC BY 4.0)

Œuvre originale : « Task Observer » par Eoghan Henn, https://rebelytics.com, dépôt
https://github.com/rebelytics/one-skill-to-rule-them-all. Licence : Creative Commons Attribution 4.0
International, https://creativecommons.org/licenses/by/4.0/ (le texte de la licence figure dans le
dépôt d'origine, fichier `LICENSE`). Aucune garantie n'est donnée par l'auteur ; l'auteur ne cautionne
pas cette adaptation.

## Modifications

- Traduction en français et réduction : seuls la liste des signaux (nouveau skill, amélioration,
  simplification) et le test de généralité d'une correction sont repris, réécrits.
- Retirés : le journal `skill-observations/` écrit dans l'espace de travail, les scripts shell
  (`session-start-scan.sh`, `skill-load.sh`, `new-observation.sh`), le balayage au démarrage de
  session, les hooks, le manifeste de plugin, la migration de format.
- Remplacés : les observations deviennent des propositions écrites dans `tmp/questions-paul.md`, jamais
  appliquées sans l'accord de Paul.
- Ajoutés : « Surcharges Encre », les déclencheurs, la recherche encadrée, le plafond de 12 skills, le
  format de rapport, le lien avec le registre et le contrôle mensuel (`tools/skills/check-upstream.mjs`).
