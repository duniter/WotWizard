# Install

To install WotWizard, see the manuals in the Help directory.

**Warning**: You need the 1.7.17 version of duniter or any later version in the series 1.7.x already installed. Versions 1.8.x don't work with WotWizard.


## Tutoriel détaillé

Vous trouverez ci-dessous les étapes détaillées pour mettre en place et configurer un serveur WotWizard.
Ces étapes ont été testées sur une installation fraiche de Debian 10.

```bash
# installation d'un noeud Duniter 1.7.21
sudo apt install unzip # dépendance nécessaire non précisée dans le paquet .deb
wget https://git.duniter.org/nodes/typescript/duniter/-/jobs/34995/artifacts/raw/work/bin/duniter-server-v1.7.21-linux-x64.deb
sudo dpkg -i duniter-server-v1.7.21-linux-x64.deb
duniter sync <noeud cible> # remplacer <noeud cible> par l'adresse du noeud de synchronisation
duniter sync-mempool <noeud cible> # si la mempool n'est pas synchronisée, la synchroniser explicitement
vi ~/.config/duniter/duniter_default/conf.json # éditer le fichier de configuration pour passer l'option wotwizard à true
duniter start # démarrer le noeud
```

Une fois le noeud bien configuré et en route, on peut passer à la suite.

```bash
# télécharger la dernière version des exécutables WotWizard 
wget https://github.com/duniter/WotWizard/releases/download/v5.1.3/wwClient 
wget https://github.com/duniter/WotWizard/releases/download/v5.1.3/wwServer
chmod u+x ww* # se donner les droits d'exécution pour les deux
./wwServeur # démarrer le serveur WotWizard, il lui faut un peu de temps pour créer sa base de données
# à l'emplacement ~/.config/duniter/duniter_default/wotwizard-export.db
# pour suivre la progression, lire les logs
tail --follow rsrc/duniter/log.txt # le dossier rsrc est créé par wwServer
# une fois que tous les blocs ont été écrits (environ 10 minutes), wwServer est prêt
# à chaque nouveau bloc, il crée un fichier temporaire updating.txt
./wwClient # on peut alors démarrer le serveur wwClient
```

Pour configurer wwClient et wwServer, éditer les fichiers de configuration présents dans rsrc.
Pour configurer un reverse proxy nginx (exemple pour wotwizard.coinduf.eu) :

```nginx
location / {
	proxy_pass http://localhost:7070;
	proxy_redirect http://localhost:7070 https://wotwizard.coinduf.eu;
}
```





