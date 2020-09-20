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
AllCertifiers	All Known Non-Revoked Certifiers:
allParameters	Blockchain Parameters
AppLimitDate	Limit Date of Membership
Availability	Availability
Bct	Blockchain Median Time
Block	Block
centralities	Centralities
Centrality	Centrality
Certifications	Certifications
certificationsFrom	Certifications From Members
certificationsTo	Certifications To Members
certLim	Certifications Limits Dates by Dates
Computation_duration	Computation Duration
day	day(s)
Delay	Delay (days)
Distance	Distance Rule
distances	Distances (rule)
Distribution	Distribution
Dossiers	Dossiers
en	English
Entry	Entry of 
Exit	Exit of 
Explorer	Explorer
explorer	Web of Trust Explorer
extCertifs	certifications outside of dossiers
FirstEntries	First Entries
FirstEntriesFlux	Flux of First Entries
FirstEntriesFluxPM	Flux of First Entries per member
Forward	Number of days until forecast date
fr	Français
Hash	Hash
history	History
hour	hour(s)
In	In
KO	KO
language	Choose Language
LCCertifiers	She was certified by 
LCComplement	This identity will lack certifications soon at the following date:
LimitsC	Certifications Limits
LimitDate	Limit Date
limitsCerts	Limit Date of Certifications
LimitsM	Limits of Memberships
limitsMember	Limit Dates of Memberships
limitsMissing	Limit Dates of not-renewed memberships
LMCertifiers	certified by
LMComplement	To stay member, this identity must renew her membership:
Losses	Losses
LossesFlux	Flux of Losses
LossesFluxPM	Flux of Losses per Member
Mean	Mean
Median	Median
Member	Member
memberIdentities	Members List
Members	Members
membersCountFlux	Evolution : Flux of Members (graphics)
membersCountFluxPM	Evolution : Flux of Members per Member (graphics)
membersCountG	Evolution : Number of members (graphics)
membersCountT	Evolution :  Number of members (list)
membersFEFlux	Evolution : Flux of First Entries (graphics)
membersFEFluxPM	Evolution : Flux of First Entries per Member (graphics)
membersFirstEntryG	Evolution : First Entries (graphics)
membersFirstEntryT	Evolution : First Entries (list)
MembersFlux	Flux of Members
MembersFluxPM	Flux of Members per Member
membersLossFlux	Evolution : Flux of Losses (graphics)
membersLossFluxPM	Evolution : Flux of Losses per Member (graphics)
membersLossG	Evolution : Losses (graphics)
membersLossT	Evolution : Losses (list)
MembersNb	Number of members
memLim	Memberships Limits Date by Date
minute	minute(s)
Missing	Excluded Identities - Needing a new application
missingIdentities	List of Excluded Identities - Needing a new application
MissingNb	Number of Excluded Identities
Missings	Excluded
month	month
Never	Never
newcomerIdentities	List of Newcomers
Newcomers	Newcomers
newcomers	Pending Dossiers
NewcomersNb	Number of Newcomers
Nickname	Nickname
no	no
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
qualities	Qualities
Quality	Quality
Revoked	Revoked
revokedIdentities	List of Revoked Identities
RevokedM	Revoked Identities
RevokedNb	Number of Revoked Identities
Revokeds	Revoked
SDev	Standard Deviation
second	second(s)
Select	Select
Sentries	Sentries
sentries	List of Sentries
SentriesNb	Number of Sentries
Sentry	Sentry
ShowFile	Dossier
SortedByCExpDates	Sorted by Expiration Dates of Certifications
Threshold	Threshold
TypeUidOrPubkey	Start of Nickname or Public Key
Utc	Blockchain Actual Time
Written_block	First Membership
wwByDate	Forecast: By Date
wwByName	Forecast: By Name
wwFile	Forecast: Preparatory File
wwMeta	Forecast: Metadata
wwPerms	Forecast: Permutations
year	year(s)
yes	yes
`
	strFr = `STRINGS

AllCertified	Toutes les identités certifiées non-révoquées :
AllCertifiers	Tous les certificateurs connus et non-révoqués :
allParameters	Paramètres de la chaîne de blocs
AppLimitDate	Date limite d'adhésion
Availability	Disponibilité
Bct	Temps médian de la chaîne de blocs
Block	Bloc
centralities	Centralités
Centrality	Centralité
Certifications	Certifications
certificationsFrom	Certifications depuis les membres
certificationsTo	Certifications vers les membres
certLim	Limites des certifications par date
Computation_duration	Durée du calcul
day	jour(s)
Delay	Durée (jours)
Distance	Règle de distance
distances	Distances (règle de)
Distribution	Distribution
Dossiers	Dossiers
en	English
Entry	Entrée de 
Exit	Sortie de 
Explorer	Explorateur
explorer	Explorateur de la toile de confiance
extCertifs	certifications hors-dossiers
FirstEntries	Premières entrées
FirstEntriesFlux	Flux des premières entrées
FirstEntriesFluxPM	Flux des premières entrées par membre
Forward	Nombre de jours jusqu'à la date de prévision 
fr	Français
Hash	Hash
history	Entrées / sorties de la toile de confiance
hour	heure(s)
In	Entrée
KO	KO
language	Choix de la langue
LCCertifiers	Elle a été certifiée par 
LCComplement	Cette identité va bientôt manquer de certifications à la date suivante :
LimitsC	Limites des certifications
LimitDate	Date limite
limitsCerts	Dates limites des certifications
LimitsM	Limites des adhésions
limitsMember	Dates limites des adhésions
limitsMissing	Dates limites des adhésions non-renouvelées
LMCertifiers	certifié(e) par
LMComplement	Pour rester membre, cette identité doit renouveler son adhésion :
Losses	Pertes
LossesFlux	Flux de pertes
LossesFluxPM	Flux de pertes par membre
Mean	Moyenne
Median	Médiane
Member	Membre
memberIdentities	Liste des membres
Members	Membres
membersCountFlux	Evolution : Flux de membres (graphique)
membersCountFluxPM	Evolution : Flux de membres par membre (graphique)
membersCountG	Evolution : Nombre de membres (graphique)
membersCountT	Evolution :  Nombre de membres (liste)
membersFEFlux	Evolution : Flux des premières entrées (graphique)
membersFEFluxPM	Evolution : Flux des premières entrées par membre (graphique)
membersFirstEntryG	Evolution : Premières entrées (graphique)
membersFirstEntryT	Evolution : Premières entrées (liste)
MembersFlux	Flux de membres
MembersFluxPM	Flux de membres par membre
membersLossFlux	Evolution : Flux de pertes (graphique)
membersLossFluxPM	Evolution : Flux de pertes par membre (graphique)
membersLossG	Evolution : Pertes (graphique)
membersLossT	Evolution : Pertes (liste)
MembersNb	Nombre de membres
memLim	Limites des adhésions par date
minute	minute(s)
Missing	Exclu(e)s en attente de réadhésion
missingIdentities	Liste des identités exclues en attente de réadhésion
MissingNb	Nombre des exclu(e)s
Missings	Exclu(e)s
month	mois
Never	Jamais
newcomerIdentities	Liste des arrivant(e)s
Newcomers	Arrivant(e)s
newcomers	dossiers en attente
NewcomersNb	Nombre des arrivant(e)s
Nickname	Pseudo
no	non
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
qualities	Qualités
Quality	Qualité
Revoked	Révoqué(e)
revokedIdentities	Liste des identités révoquées
RevokedM	Identités révoquées
RevokedNb	Nombre d'identités révoquées
Revokeds	Révoqué(e)s
SDev	Écart type
second	seconde(s)
Select	Sélectionner
Sentries	Membres référents
sentries	Liste des membres référents
SentriesNb	Nombre de membres référents
Sentry	Référent
ShowFile	Fichier
SortedByCExpDates	Tri par dates d'expiration des certifications
Threshold	Seuil
TypeUidOrPubkey	Début de pseudo ou de clef publique
Utc	Temps réel de la chaîne de blocs
Written_block	Première adhésion
wwByDate	Prévisions : par date
wwByName	Prévisions : par nom
wwFile	Prévisions : Fichier préparatoire
wwMeta	Prévisions : Métadonnées
wwPerms	Prévisions : Permutations
year	année(s)
yes	oui
`

)
