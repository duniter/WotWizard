/* 
DuniterClient: WotWizard.

Copyright (C) 2017 GérardMeunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
*/

package static

// Package strMapping implements a mapping between keys and longer strings.

var (
	
	strEn = `STRINGS

AllCertified	All Non-Revoked Certified Identities:
AllCertifiedIO	History of all sent certifications (start ↑ and end ↓):
AllCertifiers	All Known Non-Revoked Certifiers:
AllCertifiersIO	History of all received certifications (start ↑ and end ↓):
50allParameters	Informations: Blockchain Parameters
AppMLimitDate	Limit Date of Membership Renewal
AppNLimitDate	Limit Date of Membership Application
AppRLimitDate	Date of Revocation
Availability	Availability of next sent certification
Bct	Blockchain Median Time
Block	Block
Brackets	Brackets
22centralities	Properties: Centralities
centralities	Centralities
Centrality	Centrality
Certifications	Certifications
55certificationsFrom	Informations: Certifications From Members
certificationsFrom	Certifications From Members
56certificationsTo	Informations: Certifications To Members
certificationsTo	Certifications To Members
34certLim	Limits: Certifications Limits Dates by Dates
certsNb	^0 certifications
Client	Client Version: 
Computation_duration	Computation Duration
Day	Day
day	day(s)
Delay	Delay (days)
Distance	Distance Rule
distanceRule	Distance Rule: ^0%
20distances	Properties: Distances (rule)
distances	Distances (rule)
Distribution	Distribution
Dossiers	Dossiers
en	English
Entry	Entry of 
Exit	Exit of 
Explorer	Explorer
1explorer	Web of Trust Explorer
FirstEntries	First Entries
FirstEntriesFlux	Flux of First Entries
FirstEntriesFluxPM	Flux of First Entries per member
Forward	Number of days until forecast date
fr	Français
Hash	Hash
history	History
hour	hour(s)
In	In
Index	Index
index	index
intCertifs	internal certifications
KO	KO
language	Choose Language
LCCertifiers	She was certified by 
LCComplement	This identity will lack certifications soon at the following date:
LimitsC	Certifications Limits
LimitDate	Limit Date
32limitsCerts	Limits: Limit Dates of Certifications
limitsCerts	Limit Dates of Certifications
LimitsM	Limits of Memberships
30limitsMember	Limits: Limit Dates of Memberships
limitsMember	Limit Dates of Memberships
31limitsMissing	Limits: Limit Dates of not-renewed memberships
limitsMissing	Limit Dates of not-renewed memberships
LMCertifiers	certified by
LMComplement	To stay member, this identity must renew her membership:
Losses	Losses
LossesFlux	Flux of Losses
LossesFluxPM	Flux of Losses per Member
Mean	Mean
Median	Median
MEMBER	MEMBER
53memberIdentities	Informations: Members List
Members	Members
404membersCountFlux	Evolution: Flux of Members (graphics)
405membersCountFluxPM	Evolution: Flux of Members per Member (graphics)
403membersCountG	Evolution: Number of members (graphics)
400membersCountT	Evolution:  Number of members (list)
407membersFEFlux	Evolution: Flux of First Entries (graphics)
408membersFEFluxPM	Evolution: Flux of First Entries per Member (graphics)
406membersFirstEntryG	Evolution: First Entries (graphics)
401membersFirstEntryT	Evolution: First Entries (list)
MembersFlux	Flux of Members
MembersFluxPM	Flux of Members per Member
410membersLossFlux	Evolution: Flux of Losses (graphics)
411membersLossFluxPM	Evolution: Flux of Losses per Member (graphics)
409membersLossG	Evolution: Losses (graphics)
402membersLossT	Evolution: Losses (list)
MembersNb	Number of members
33memLim	Limits: Memberships Limits Date by Date
minApplicationDate	Wait at least two months after the last application (^0)
minute	minute(s)
MISSING	MISSING
Missing	Excluded Identities - Needing a new application
52missingIdentities	Informations: List of Excluded Identities - Needing a new application
MissingNb	Number of Excluded Identities
Missings	Excluded
month	month
Never	Never
NEWCOMER	NEWCOMER
54newcomerIdentities	Informations: List of Newcomers
Newcomers	Newcomers
newcomers	Pending Dossiers
NewcomersNb	Number of Newcomers
Nickname	Nickname
no	no
NotMembers	Not Members
OK	OK
Out	Out
Parameters	Parameters
pending	pending new application
Permutations	Permutations
permutations	permutations
PermutationsNb	Number of Permutations: 
PresentCertified	Currently or Soon (°) Certified Identities:
PresentCertifiers	Current or Coming (°) Certifiers:
Proba	Probability
Pubkey	Public Key
21qualities	Properties: Qualities
qualities	Qualities
Quality	Quality
requiredCertsNb	^0 certifications, ^1 needed for distance rule
REVOKED	REVOKED
Revoked	Revoked
51revokedIdentities	Informations: List of Revoked Identities
RevokedM	Revoked Identities
RevokedNb	Number of Revoked Identities
Revokeds	Revoked
SDev	Standard Deviation
second	second(s)
Select	Select
SentCertsHistory	Sent Certifications History
60sentCertsHistory	Statistics: Sent Certifications History
Sentries	Sentries
57sentries	Informations : List of Sentries
SentriesNb	Number of Sentries
Sentry	Sentry
Server	Server Version: 
ShowFile	Dossier
SortedByCExpDates	Sorted by Expiration Dates of Certifications
SortedByCExpDatesL	Sorted by Expiration Dates of Certifications (→: limit date)
Status	Status
Threshold	Threshold
TypeUidOrPubkey	Start of Nickname or Public Key
Utc	Blockchain Actual Time
WarnNum	Number of Warnings:
Written_block	First Membership
wwByDate	Sorted by Dates
wwByName	Sorted by Names
01wwFile	Forecast: Preparatory File
wwMeta	Metadata
02wwPerms	Forecast: Permutations
00wwView	ForeCast: WotWizard View
year	year(s)
yes	yes
`
	strFr = `STRINGS

AllCertified	Toutes les identités certifiées non-révoquées :
AllCertifiedIO	Débuts (↑) et fins (↓) de validité de toutes les certifications émises :
AllCertifiers	Tous les certificateurs connus et non-révoqués :
AllCertifiersIO	Débuts (↑) et fins (↓) de validité de toutes les certifications reçues :
50allParameters	Informations : Paramètres de la chaîne de blocs
AppMLimitDate	Date limite de réadhésion
AppNLimitDate	Date limite de la demande d'adhésion
AppRLimitDate	Date de révocation
Availability	Disponibilité du prochain envoi de certification
Bct	Temps médian de la chaîne de blocs
Block	Bloc
Brackets	Tranches
22centralities	Propriétés : Centralités
centralities	Centralités
Centrality	Centralité
Certifications	Certifications
55certificationsFrom	Informations : Certifications depuis les membres
certificationsFrom	Certifications depuis les membres
56certificationsTo	Informations : Certifications vers les membres
certificationsTo	Certifications vers les membres
34certLim	Limites : Limites des certifications par date
certsNb	^0 certifications
Client	Version client : 
Computation_duration	Durée du calcul
Day	Jour
day	jour(s)
Delay	Durée (jours)
Distance	Règle de distance
distanceRule	Règle de distance: ^0%
20distances	Propriétés : Distances (règle de)
distances	Distances (règle de)
Distribution	Distribution
Dossiers	Dossiers
en	English
Entry	Entrée de 
Exit	Sortie de 
Explorer	Explorateur
1explorer	Explorateur de la toile de confiance
FirstEntries	Premières entrées
FirstEntriesFlux	Flux des premières entrées
FirstEntriesFluxPM	Flux des premières entrées par membre
Forward	Nombre de jours jusqu'à la date de prévision 
fr	Français
Hash	Hash
history	Entrées / sorties de la toile de confiance
hour	heure(s)
In	Entrée
Index	Menu
index	menu
intCertifs	certifications internes
KO	KO
language	Choix de la langue
LCCertifiers	Elle a été certifiée par 
LCComplement	Cette identité va bientôt manquer de certifications à la date suivante :
LimitsC	Limites des certifications
LimitDate	Date limite
32limitsCerts	Limites : Dates limites des certifications
limitsCerts	Dates limites des certifications
LimitsM	Limites des adhésions
30limitsMember	Limites : Dates limites des adhésions
limitsMember	Dates limites des adhésions
31limitsMissing	Limites : Dates limites des adhésions non-renouvelées
limitsMissing	Dates limites des adhésions non-renouvelées
LMCertifiers	certifié(e) par
LMComplement	Pour rester membre, cette identité doit renouveler son adhésion :
Losses	Pertes
LossesFlux	Flux de pertes
LossesFluxPM	Flux de pertes par membre
Mean	Moyenne
Median	Médiane
MEMBER	MEMBRE
53memberIdentities	Informations : Liste des membres
Members	Membres
404membersCountFlux	Evolution : Flux de membres (graphique)
405membersCountFluxPM	Evolution : Flux de membres par membre (graphique)
403membersCountG	Evolution : Nombre de membres (graphique)
400membersCountT	Evolution :  Nombre de membres (liste)
407membersFEFlux	Evolution : Flux des premières entrées (graphique)
408membersFEFluxPM	Evolution : Flux des premières entrées par membre (graphique)
406membersFirstEntryG	Evolution : Premières entrées (graphique)
401membersFirstEntryT	Evolution : Premières entrées (liste)
MembersFlux	Flux de membres
MembersFluxPM	Flux de membres par membre
410membersLossFlux	Evolution : Flux de pertes (graphique)
411membersLossFluxPM	Evolution : Flux de pertes par membre (graphique)
409membersLossG	Evolution : Pertes (graphique)
402membersLossT	Evolution : Pertes (liste)
MembersNb	Nombre de membres
33memLim	Limites : Limites des adhésions par date
minApplicationDate	Au moins deux mois d'attente après la dernière adhésion (^0)
minute	minute(s)
MISSING	EXCLU(E)
Missing	Exclu(e)s en attente de réadhésion
52missingIdentities	Informations : Liste des identités exclues en attente de réadhésion
MissingNb	Nombre des exclu(e)s
Missings	Exclu(e)s
month	mois
Never	Jamais
NEWCOMER	ARRIVANT(E)
54newcomerIdentities	Informations : Liste des arrivant(e)s
Newcomers	Arrivant(e)s
newcomers	dossiers en attente
NewcomersNb	Nombre des arrivant(e)s
Nickname	Pseudo
no	non
NotMembers	Non membres
OK	OK
Out	Sortie
Parameters	Paramètres
pending	En cours de réadhésion
Permutations	Permutations
permutations	permutations
PermutationsNb	Nombre de permutations : 
PresentCertified	Identités actuellement ou prochainement (°) certifiées :
PresentCertifiers	Actuels ou prochains (°) certificateurs :
Proba	Probabilité
Pubkey	Clef publique
21qualities	Propriétés : Qualités
qualities	Qualités
Quality	Qualité
requiredCertsNb	^0 certifications, ^1 nécessaires pour la règle de distance
REVOKED	RÉVOQUÉ(E)
Revoked	Révoqué(e)
51revokedIdentities	Informations : Liste des identités révoquées
RevokedM	Identités révoquées
RevokedNb	Nombre d'identités révoquées
Revokeds	Révoqué(e)s
SDev	Écart type
second	seconde(s)
Select	Sélectionner
SentCertsHistory	Évolution des certifications émises
60sentCertsHistory	Statistiques : Évolution des certifications émises
Sentries	Membres référents
57sentries	Informations : Liste des membres référents
SentriesNb	Nombre de membres référents
Sentry	Référent
Server	Version serveur : 
ShowFile	Fichier
SortedByCExpDates	Tri par dates d'expiration des certifications
SortedByCExpDatesL	Tri par dates d'expiration des certifications (→ : date limite)
Status	Statut
Threshold	Seuil
TypeUidOrPubkey	Début de pseudo ou de clef publique
Utc	Temps réel de la chaîne de blocs
WarnNum	Nombre d'alertes :
Written_block	Première adhésion
wwByDate	Tri par dates
wwByName	Tri par noms
01wwFile	Prévisions : Fichier préparatoire
wwMeta	Métadonnées
02wwPerms	Prévisions : Permutations
00wwView	Prévisions: Fenêtre WotWizard
year	année(s)
yes	oui
`

)
