/*
util: Set of tools.

Copyright (C) 2001-2020 Gérard Meunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA 02111-1307, USA.
*/

package static

var (
	
	typeSystem = `
# WotWizard GraphQL TypeSystem

type Query {
	
	"'identities' lists all identities whose status is 'status' and whose uids is between 'start' (included) and 'end' (excluded), in increasing order and sorted by 'sortedBy'; if 'start' is absent or null, the list starts at the beginning, and stops at the end if 'end' is absent or null"
	identities (status: Identity_Status! = MEMBER, sortedBy: Identity_Order! = UID, start: String! = "", end: String! = ""): [Identity!]!
	
	"'idSearch' displays the list of identities whose pseudos or public keys begin with 'with.hint' and whose status is in 'with.status_list'."
	idSearch (with: IdSearchInput! = {hint: "", status_list: [REVOKED, MISSING, MEMBER, NEWCOMER]}): IdSearchOutput!
	
	"'idFromHash' retreives the 'Identity' whose hash is 'hash'; it returns null if this identity doesn't exist"
	idFromHash (hash: Hash!): Identity
	
	"Threshold for numbers of sent and received certifications to become sentry"
	sentryTreshold: Int!
	
	"List of sentries, sorted by increasing uids"
	sentries: [Identity!]!
	
	"Present block"
	now: Block!

	"'wwFile' displays the WotWizard file, complete if 'full', or else with Dossier(s) containing at least 'Query.parameter(name: sigQty)' certifications only"
	wwFile (full: Boolean! = false): File!
	
	"'wwResult' displays the content of the WotWizard window"
	wwResult: WWResult!
	
	"'memEnds' displays the list of members who are about to loose their memberships, in the order of event dates (bct); 'startFromNow' gives the period before the beginning of the list (0 if absent or null) , and 'period' gives the period covered by the list (infinite if absent or null)"
	memEnds (startFromNow: Int64, period: Int64): [Identity!]!
	
	"'missEnds' displays the list of identities who are MISSING and about to be revoked, in the order of event dates (bct); 'startFromNow' gives the period before the beginning of the list (0 if absent or null) , and 'period' gives the period covered by the list (infinite if absent or null)"
	missEnds (startFromNow: Int64, period: Int64): [Identity!]!
	
	"'certEnds' displays the list of identities who are MEMBER or (possibly) MISSING and about to loose their 'ParameterName.sigQty'th oldest certification, in the order of event dates (bct); 'startFromNow' gives the period before the beginning of the list (0 if absent or null) , and 'period' gives the period covered by the list (infinite if absent or null)"
	certEnds (startFromNow: Int64, period: Int64, missingIncluded: Boolean! = true): [Identity!]!
	
	"'countMin' gives the first block of the blockchain"
	countMin: Block!
	
	"'countMax' gives the last present block of the blockchain"
	countMax: Block!
	
	"'membersCount' displays the number of active members, sorted by dates (utc0) of events (in or out the wot); if 'start' is absent or null, the display starts at 'countMin', and ends at 'countMax' if 'end' is absent or null"
	membersCount (start: Int64, end: Int64): [Event!]!
	
	"'membersFlux' displays the flux of active members by <timeUnit (s)>; if 'start' is absent or null, the display starts at 'countMin', and ends at 'countMax' if 'end' is absent or null"
	membersFlux (start: Int64, end: Int64, timeUnit: Int64! = 2629800): [FluxEvent!]!
	
	"'membersFluxPM' displays the flux of active members by <timeUnit (s)> and by member; if 'start' is absent or null, the display starts at 'countMin', and ends at 'countMax' if 'end' is absent or null"
	membersFluxPM (start: Int64, end: Int64, timeUnit: Int64! = 2629800): [FluxEvent!]!
	
	"'fECount' displays the number of first entries into the wot, sorted by dates (utc0) of events (entries); if 'start' is absent or null, the display starts at 'countMin', and ends at 'countMax' if 'end' is absent or null"
	fECount (start: Int64, end: Int64): [Event!]!
	
	"'fEFlux' displays the flux of first entries by <timeUnit (s)>; if 'start' is absent or null, the display starts at 'countMin', and ends at 'countMax' if 'end' is absent or null"
	fEFlux (start: Int64, end: Int64, timeUnit: Int64! = 2629800): [FluxEvent!]!
	
	"'fEFluxPM' displays the flux of first entries by <timeUnit (s)> and by member; if 'start' is absent or null, the display starts at 'countMin', and ends at 'countMax' if 'end' is absent or null"
	fEFluxPM (start: Int64, end: Int64, timeUnit: Int64! = 2629800): [FluxEvent!]!
	
	"'lossCount' displays the number of members exiting the wot, minus the number of reentries (losses), sorted by dates (utc0) of events (in or out the wot); if 'start' is absent or null, the display starts at 'countMin', and ends at 'countMax' if 'end' is absent or null"
	lossCount (start: Int64, end: Int64): [Event!]!
	
	"'lossFlux' displays the flux of losses by <timeUnit (s)>; if 'start' is absent or null, the display starts at 'countMin', and ends at 'countMax' if 'end' is absent or null"
	lossFlux (start: Int64, end: Int64, timeUnit: Int64! = 2629800): [FluxEvent!]!
	
	"'lossFluxPM' displays the flux of losses by <timeUnit (s)> and by member; if 'start' is absent or null, the display starts at 'countMin', and ends at 'countMax' if 'end' is absent or null"
	lossFluxPM (start: Int64, end: Int64, timeUnit: Int64! = 2629800): [FluxEvent!]!
	
	"'allParameters' displays all parameters of the money"
	allParameters: [Parameter!]!
	
	"'parameter' displays the parameter of the money whose name is 'name''"
	parameter (name: ParameterName): Parameter

} #Query

type Mutation {
	
	"'stopSubscription' erases the subscription whose name is 'name', which sends results at address 'returnAddr'; 'varVals' is a JSON object whose fields keys are the names of the variables (without '$') used in the subscription and whose fields values are their values"
	stopSubscription (returnAddr: String!, name: String!, varVals: String): Void
	
	"changeDifferParams modifies two parameters of the differentiation process used in 'Query.membersFlux', 'Query.membersFluxPM', 'Query.fEFlux', 'Query.fEFluxPM', 'Query.lossFlux' and 'Query.lossFluxPM': 'pointsNb', the number of points over which the filter (Savitzky-Golay filter) is calculated, and 'degree',  the degree of the used polynomial (usually 2 or 4); do not change a parameter if absent or null. It returns the previous values of these two parameters."
	changeDifferParams (pointsNb: Int, degree: Int): DifferParams!

} #Mutation

type Subscription {
	
	"'now' installs a subscription for the update of 'Query.now' at every new block"
	now: Block!

	"'wwFile' installs a subscription for the update of 'Query.wwFile' at every new block"
	wwFile (full: Boolean! = false): FileS!
	
	"'wwResult' installs a subscription for the update of 'Query.wwResult' at every new block"
	wwResult: WWResultS!
	
	"'memEnds' installs a subscription for the update of 'Query.memEnds' at every new block"
	memEnds (startFromNow: Int64, period: Int64): [Identity!]!
	
	"'missEnds' installs a subscription for the update of 'Query.missEnds' at every new block"
	missEnds (startFromNow: Int64, period: Int64): [Identity!]!
	
	"'certEnds' installs a subscription for the update of 'Query.certEnds' at every new block"
	certEnds (startFromNow: Int64, period: Int64, missingIncluded: Boolean! = true): [Identity!]!

} #Subscription

"WoT identity"
type Identity {
	
	"Public key"
	pubkey: Pubkey!
	
	"Pseudo"
	uid: String!
	
	"Hash"
	hash: Hash!
	
	"Status: REVOKED, MISSING, MEMBER or NEWCOMER"
	status: Identity_Status!
	
	"Identity waiting for new membership (MISSING, MEMBER or NEWCOMER; false for REVOKED)"
	membership_pending: Boolean!
	
	"Block of new membership application; null if not membership_pending"
	membership_pending_block: Block
	
	"Limit date of new membership application; null if not membership_pending"
	membership_pending_limitDate: Int64
	
	"Block where the identity is written (entry into the web of trust); null for NEWCOMER"
	id_written_block: Block
	
	"Block of last membership application (joiners, actives, leavers), null for NEWCOMER"
	lastApplication: Block
	
	"Limit date of the membership application; null for REVOKED; limit date before revocation for MISSING"
	limitDate: Int64
	
	"Member is leaving? null for REVOKED or NEWCOMER"
	isLeaving: Boolean
	
	"Is a sentry? null if not MEMBER"
	sentry: Boolean
	
	"Active received certifications, sorted by increasing pubkeys"
	received_certifications: Received_Certifications!
	
	"Active sent certifications, sorted by increasing pubkeys"
	sent_certifications: [Certification!]!
	
	"All certifiers, old or present, but not REVOKED (empty list for NEWCOMER)"
	all_certifiers: [Identity!]!
	
	"All certified identities, old or present, but not REVOKED (empty list for NEWCOMER)"
	all_certified: [Identity!]!
	
	"State of the identity's distance rule"
	distance: Distance!
	
	"Identity's quality (percent)"
	quality: Float!
	
	"Identity's degree of centrality (percent)"
	centrality: Float!
	
	"History of identity's entries into and exits out of the WoT (empty list for NEWCOMER)"
	history: [HistoryEvent!]!
	
	"Minimum date of next sent certification; null if not MEMBER"
	minDate: Int64
	
	"Minimum date of next sent certification is passed; null if not MEMBER"
	minDatePassed: Boolean

} #Identity

"Status of an identity in WoT"
enum Identity_Status {
	
	"The identity has been revoked"
	REVOKED
	
	"The identity is no more member but not revoked yet"
	MISSING
	
	"The identity is a member of the WoT"
	MEMBER
	
	"Newcomer waiting for membership"
	NEWCOMER

} #Identity_Status

"Sorting order"
enum Identity_Order {
	
	"Sorting by uid"
	UID
	
	"Sorting by pubkey"
	PUBKEY

} #Identity_Order

"Used by 'Query.idSearch'"
input IdSearchInput {
	
	"Prefix of searched keys (uid or pubkey)"
	hint: String! = ""
	
	"List of searched statuses"
	status_list: [Identity_Status!]! = [REVOKED, MISSING, MEMBER, NEWCOMER]
	
} #IdSearchInput

"Result of 'Query.idSearch'"
type IdSearchOutput {
	
	"Number of REVOKED identities corresponding to 'IdSearchInput.hint'"
	revokedNb: Int!
	
	"Number of MISSING identities corresponding to 'IdSearchInput.hint'"
	missingNb: Int!
	
	"Number of MEMBER identities corresponding to 'IdSearchInput.hint'"
	memberNb: Int!
	
	"Number of NEWCOMER identities corresponding to 'IdSearchInput.hint'"
	newcomerNb: Int!
	
	"All identities corresponding to 'IdSearchInput'"
	ids: [Identity!]!
	
} #IdSearchOutput

"Certifications received by an identity"
type Received_Certifications {
	
	"List of all valid received certifications"
	certifications: [Certification!]!
	
	"Limit date of the 'ParameterName.sigQty'th oldest received certification; or null if less than 'ParameterName.sigQty' certifications received"
	limit: Int64
	
} #received_Certifications

"Certification sent by 'from' and received by 'to'"
type Certification {
	
	"Sender"
	from: Identity!
	
	"Receiver"
	to: Identity!
	
	"Is certification in sandbox?"
	pending: Boolean!
	
	"Registration block"
	block: Block
	
	"Limit date (bct) of validity"
	expires_on: Int64!
	
} #Certification

"Result of distance rule evaluation"
type Distance {
	
	"Proportion of sentries reached in 'ParameterName.stepMax' steps or less (percent)"
	value: Float!
	
	"Is 'value' greater than 'ParameterName.xpercent' or equal?"
	dist_ok: Boolean!
	
} #Distance

"History of entries into the WoT and exits of an identity"
type HistoryEvent {
	
	"Entry?"
	in: Boolean!
	
	"Registration block"
	block: Block!
	
} #HistoryEvent

"Number & dates of a block"
type Block {
	
	"Block number"
	number: Int!
	
	"Blockchain time"
	bct: Int64!
	
	"UTC+0 real time"
	utc0: Int64!

} #Block

"Differentiation filter parameters"
type DifferParams {

	"Number of points used by the filter"
	pointsNb: Int!
	
	"Degree of polynomial used by the filter"
	degree: Int!
	
} #DifferParams

"Set of internal certifications and membership application dossiers available in sandbox"
interface File {

	"List of internal certifications and membership application dossiers"
	certifs_dossiers: [CertifOrDossier!]!
	
	"Number od dossiers"
	dossiers_nb: Int!
	
	"Number of internal certifications"
	certifs_nb: Int!

} #File

"Set of internal certifications and membership application dossiers available in sandbox; dated"
type FileS implements File {
	
	"Present block"
	now: Block!

	"List of internal certifications and membership application dossiers"
	certifs_dossiers: [CertifOrDossier!]!
	
	"Number od dossiers"
	dossiers_nb: Int!
	
	"Number of internal certifications"
	certifs_nb: Int!

} #FileS

"Internal certification or membership application dossier"
union CertifOrDossier = MarkedDatedCertification | MarkedDossier

"Inserted type used to distinguish 'DatedCertification'(s) and 'Dossier'(s) in 'CertifOrDossier'"
type MarkedDatedCertification {
	datedCertification: DatedCertification!
} #MarkedDatedCertification

"Inserted type used to distinguish 'DatedCertification'(s) and 'Dossier'(s) in 'CertifOrDossier'"
type MarkedDossier {
	dossier: Dossier!
} #MarkedDossier

"Certification in a 'File'"
type DatedCertification {
	
	certification: Certification!
	
	"Date of availability"
	date: Int64!

} #DatedCertification

"Newcomer's membership application dossier"
type Dossier {
	
	newcomer: Identity!
	
	"Minimum number of certifications needed to fulfill the distance rule"
	main_certifs: Int!
	
	"External certifications"
	certifications: [DatedCertification!]!
	
	"'ParameterName.msPeriod' after the last membership application (or 0 if none)"
	minDate: Int64!
	
	"Date of availability"
	date: Int64!
	
	"Expiration date"
	limit: Int64!

} #Dossier

"Result of 'Query.wwResult'"
interface WWResult {
	
	"Total time of computation, 'File' included"
	computation_duration: Int!
	
	"Number of permutations used; this number may be very big"
	permutations_nb: Int!
	
	"Number of NEWCOMER(s)' dossiers"
	dossiers_nb: Int!
	
	"Number of internal certifications"
	certifs_nb: Int!
	
	"'permutations' displays the list of WotWizard permutations; their number may be very big"
	permutations: [WeightedPermutation!]!
	
	"Forecasts of NEWCOMER(s)' entries, sorted by dates of entries"
	forecastsByDates: [Forecast!]!
	
	"Forecasts of entries of the NEWCOMER(s) whose uid(s) begin with the 'with' parameter (or of all NEWCOMER(s) if 'with' is absent or null); the selection is not case sensitive"
	forecastsByNames (with: String! = ""): [Forecast!]!
	
} #WWResult

"Result of 'Subscription.wwResult; dated'"
type WWResultS implements WWResult {
	
	"Present block"
	now: Block!
	
	"Total time of computation, 'File' included"
	computation_duration: Int!
	
	"Number of permutations used; this number may be very big"
	permutations_nb: Int!
	
	"Number of NEWCOMER(s)' dossiers"
	dossiers_nb: Int!
	
	"Number of internal certifications"
	certifs_nb: Int!
	
	"'permutations' displays the list of WotWizard permutations; their number may be very big"
	permutations: [WeightedPermutation!]!
	
	"Forecasts of NEWCOMER(s)' entries, sorted by dates of entries"
	forecastsByDates: [Forecast!]!
	
	"Forecasts of entries of the NEWCOMER(s) whose uid(s) begin with the 'with' parameter (or of all NEWCOMER(s) if 'with' is absent or null); the selection is not case sensitive"
	forecastsByNames (with: String! = ""): [Forecast!]!
	
} #WWResultS

"A permutation weighted by a probability"
type WeightedPermutation {
	
	"Probability of occurrence"
	proba: Float!
	
	"Ordered list of NEWCOMER(s)' entries"
	permutation: [PermutationElem!]!
	
} #WeightedPermutation

"An expected NEWCOMER's entry"
type PermutationElem {
	
	id: Identity!
	
	"Expected date of entry"
	date: Int64!
	
	"The expected date of entry may be later than 'date' (the computing was interrupted by lack of memory space)"
	after: Boolean!
	
} #PermutationElem

"Forecast of a NEWCOMER's entry"
type Forecast {
	
	id: Identity!
	
	"Expected date of entry"
	date: Int64!
	
	"The expected date of entry may be later than 'date' (the computing was interrupted by lack of memory space)"
	after: Boolean!
	
	"Probability of the forecast"
	proba: Float!
	
} #Forecast

"Entry or exit of an identity"
type EventId {
	
	id: Identity!
	
	"Entry or exit; true if entry"
	inOut: Boolean!
	
} #EventId

"Entries and exits of identities happening in a block"
type Event {
	
	"List of concerned identities"
	idList: [EventId!]!
	
	"Block where the event happens"
	block: Block!
	
	"Number of concerned identities in the WoT after the event"
	number: Int!
	
} #Event

"An event with non-integer value, typically a flux of entries/exits"
type FluxEvent {
	
	"Block where the event happens"
	block: Block!
	
	"Value of the flux at the event"
	value: Float!
} #FluxEvent

"A parameter of the money"
type Parameter {
	
	name: ParameterName!
	
	par_type: ParameterType!
	
	value: Number!
	
	comment: String
	
} #Parameter

enum ParameterType {
	INTEGER
	FLOAT
	DURATION
	DATE
} #ParameterType

enum ParameterName {
	
	"The relative growth of the UD every [dtReeval] period"
	c
	
	"Time period between two UD"
	dt
	
	"UD(0), i.e. initial Universal Dividend"
	ud0
	
	"Minimum delay between two certifications of a same issuer"
	sigPeriod
	
	"Maximum quantity of active certifications made by member"
	sigStock
	
	"Maximum delay a certification can wait before being expired for non-writing"
	sigWindow
	
	"Maximum age of an active certification"
	sigValidity
	
	"Minimum delay before replaying a certification"
	sigReplay
	
	"Minimum quantity of signatures to be part of the WoT"
	sigQty
	
	"Maximum delay an identity can wait before being expired for non-writing"
	idtyWindow
	
	"Maximum delay a membership can wait before being expired for non-writing"
	msWindow
	
	"Minimum delay between 2 memberships of a same issuer"
	msPeriod
	
	"Minimum percent of sentries to reach to match the distance rule"
	xpercent
	
	"Maximum age of an active membership"
	msValidity
	
	"Maximum distance between a WOT member and [xpercent] of sentries"
	stepMax
	
	"Number of blocks used for calculating median time"
	medianTimeBlocks
	
	"The average time for writing 1 block (wished time)"
	avgGenTime
	
	"The number of blocks required to evaluate again PoWMin value"
	dtDiffEval
	
	"The proportion of calculating members not excluded from the proof of work"
	percentRot
	
	"Time of first UD"
	udTime0
	
	"Time of first reevaluation of the UD"
	udReevalTime0
	
	"Time period between two re-evaluation of the UD"
	dtReeval
	
	"Maximum delay a transaction can wait before being expired for non-writing"
	txWindow

} #ParameterName

"64 bits signed integer"
scalar Int64

"Avatar of String"
scalar Hash

"Avatar of String"
scalar Pubkey

"Empty result"
scalar Void

"Int, Int64 or Float"
scalar Number
`

)
