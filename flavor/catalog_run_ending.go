package flavor

// RunEnding is the end of a run: one line logged at every prestige, after the
// Last Passage verdict (when there is one) and before the new run begins.
//
// The request shape:
//
//	flavor.Request{
//		Moment:  flavor.RunEnding,
//		Age:     age,    // the age the run ended in
//		Subject: h.Name, // the age's harbinger, Cosmic Era only; "" before it
//	}
//
// A full run's prestige comes from the Modern Age on, and the pools those
// runs reach are below. (Prestige opens at the Medieval Age, but a taste,
// from the Medieval to the Atomic Age, reaches only the era-neutral pool for
// now: pools of its own are a follow-up.)
//
//   - the Modern to the Space Age: a plain register, a civilization winding
//     down. Offices empty, the last tram runs, the orbital yards shut. Nobody
//     warns of anything; the harbinger of those ages is not in the room.
//   - the Cosmic Era: the age's harbinger is there at the end (the Distress
//     Beacon, the Elder Relay, your future self, your unmade self), and the
//     prose drifts into the late-game dread the harbinger catalog already has.
//     These ages reach none of the plain pools, only the small era-neutral one,
//     so the figure's voice dominates the draw.
//
// The era-neutral pool exists so a zero Request, and a request for an age that
// can never prestige, still produces a sentence. Like the harbinger Moments,
// RunEnding fires rarely (once a run), so it is sized by its own floor rather
// than TestSkeletonFloors: every age that can prestige reaches a pool wider
// than the Stream's window. The same no-digit rule applies: the game prints the
// points.

// runEndPlainAges are the ages before the Cosmic Era from which a full run
// prestiges: the Modern Age through the Space Age.
var runEndPlainAges = ages("modern_age", "information_age", "digital_age", "cyberpunk_age", "fusion_age", "space_age")

// runEndCosmicAges are the Cosmic Era's ages, whose passage is prestige itself.
var runEndCosmicAges = ages("interstellar_age", "galactic_age", "quantum_age", "transcendent_age")

func runEndingTemplates() []tmpl {
	out := pool("run_end_any", erasAny, runEndAny)
	out = append(out, agePool("run_end_plain", runEndPlainAges, runEndPlain)...)
	out = append(out, pool("run_end_industrial", erasIndustrial, runEndIndustrial)...)
	out = append(out, pool("run_end_digital", erasDigital, runEndDigital)...)
	out = append(out, agePool("run_end_space", ages("space_age"), runEndSpace)...)
	out = append(out, agePool("run_end_cosmic", runEndCosmicAges, runEndCosmic)...)
	out = append(out, agePools("run_end_age", runEndAges)...)
	return out
}

// runEndAny fits every age: no era-coded noun at all.
var runEndAny = []skel{
	{Text: "Everyone came outside", Reg: rPlain, Topic: "town"},
	{Text: "Names were written down in the order people came forward to give them", Reg: rPlain, Topic: "name"},
	{Text: "Debts were forgiven, since there was no longer anyone to collect them", Reg: rWry, Topic: "money"},
	{Text: "The grumbling was that it had come before anything was finished", Reg: rWry, Form: fComplaint, Topic: "time"},
	{Text: "Whatever could be carried was handed to whoever could carry it, and what could not be carried was left where it stood with a note saying whose it had been", Reg: rPlain, Topic: "haul"},
	{Text: "Doors were left open", Reg: rPlain, Topic: "building"},
	{Text: "Food was shared out until it was gone", Reg: rPlain, Topic: "food"},
	{Text: "Most people wanted to be with their families, and were", Reg: rPlain, Topic: "family"},
}

// runEndPlain is the winding-down voice from the Modern to the Space Age. It
// spans three era buckets, so it names nothing any one of them owns.
var runEndPlain = []skel{
	{Text: "People put down what they were holding", Reg: rPlain, Topic: "work"},
	{Text: "The children were told a shorter version", Reg: rWry, Topic: "family"},
	{Text: "A meal went cold on a table while the family that cooked it stood outside looking up", Reg: rPlain, Topic: "food"},
	{Text: "The oldest people walked out to look at the end for themselves and came back without much to say", Reg: rPlain, Topic: "people"},
	{Text: "Nothing new was started", Reg: rPlain, Topic: "time"},
	{Text: "Last instructions were posted where everyone could read them", Reg: rPlain, Form: fNotice, Topic: "message"},
	{Text: "Kept, one of everything, and a copy of the names", Reg: rPlain, Form: fLedger, Topic: "count"},
	{Text: "People said it had been a good long life for a people, and most of them meant it", Reg: rWry, Form: fOverheard, Topic: "rumour"},
	{Text: "The last thing anyone agreed on was what should be left behind for whoever came next, and after so many years of arguing about everything else they settled it in an afternoon", Reg: rWry, Topic: "argument"},
	{Text: "In the end they spent the last of it like a holiday they knew would not come again, visiting neighbours they had never once spoken to in all the years before and standing about in doorways", Reg: rPlain, Topic: "people"},
	{Text: "A man went round at last to make an apology he had owed his brother since they were boys, and was forgiven at once, and was annoyed about it", Reg: rJoke, Topic: "family"},
	{Text: "The complaint going round is that the end came on a weekday", Reg: rJoke, Form: fComplaint, Topic: "time"},
}

// runEndIndustrial is the Modern Age city: mills, trams, the town hall.
var runEndIndustrial = []skel{
	{Text: "The factories let their last shift go early", Reg: rPlain, Topic: "work"},
	{Text: "The chimneys stopped smoking", Reg: rPlain, Topic: "building"},
	{Text: "The power station was run down one boiler at a time", Reg: rPlain, Topic: "machine"},
	{Text: "Offices emptied by noon", Reg: rPlain, Topic: "town"},
	{Text: "The newspapers printed a final edition", Reg: rPlain, Topic: "paper"},
	{Text: "The last tram of the night went round its whole route with every light on and not one passenger aboard", Reg: rPlain, Topic: "machine"},
	{Text: "Final pay packets, issued in full", Reg: rPlain, Form: fLedger, Topic: "money"},
	{Text: "The works will not reopen on Monday", Reg: rPlain, Form: fNotice, Topic: "work"},
	{Text: "The union complains that the end came without the statutory notice", Reg: rWry, Form: fComplaint, Topic: "authority"},
	{Text: "They are saying in the pubs that it was always going to end like this, and they said it last week too", Reg: rWry, Form: fOverheard, Topic: "rumour"},
	{Text: "Families had their photographs taken in front of their houses", Reg: rPlain, Topic: "family"},
	{Text: "On the final broadcast a dance band played to an empty studio", Reg: rPlain, Topic: "noise"},
	{Text: "Men from the rolling mill walked home together in their work clothes for the last time, and stood at the corner a while before anyone went in", Reg: rPlain, Topic: "people"},
	{Text: "At the town hall the typists worked through the evening copying out the deeds and the registers and the lists of the dead, so that whoever came afterwards would find the whole of it in order and know who had lived there", Reg: rPlain, Topic: "paper"},
	{Text: "The tax office sent out its final demands on schedule", Reg: rJoke, Topic: "money"},
}

// runEndDigital is the Information to the Fusion Age: feeds, phones, arcades.
var runEndDigital = []skel{
	{Text: "The network went down region by region", Reg: rPlain, Topic: "machine"},
	{Text: "Phones stopped ringing", Reg: rPlain, Topic: "noise"},
	{Text: "Most of the final posts were photographs of the sky", Reg: rPlain, Topic: "message"},
	{Text: "Final balances, frozen at midnight", Reg: rPlain, Form: fLedger, Topic: "money"},
	{Text: "Customers complain that the promised update never shipped", Reg: rJoke, Form: fComplaint, Topic: "trade"},
	{Text: "Staff may keep their company phones", Reg: rWry, Form: fNotice, Topic: "work"},
	{Text: "Word on the feeds is that the corporations knew and sold tickets", Reg: rWry, Form: fOverheard, Topic: "rumour"},
	{Text: "The delivery drones were called home and parked in rows on the landing pads, where they sat blinking", Reg: rPlain, Topic: "machine"},
	{Text: "The fusion plants were brought down to idle one after another, and the city dimmed in stages over three nights", Reg: rPlain, Topic: "town"},
	{Text: "In the arcades the kids who had played one game every night for years played it one last time together, and when it ended they left the cabinets switched on, cycling their demo screens to an empty room", Reg: rPlain, Topic: "people"},
	{Text: "Screens across the city went dark one block at a time", Reg: rPlain, Topic: "building"},
	{Text: "Every server farm was left running for as long as the power held, humming to itself in the dark", Reg: rPlain, Topic: "machine"},
	{Text: "People stayed logged in to the end, scrolling back through years of their own posts and pictures", Reg: rPlain, Topic: "people"},
	{Text: "All personal files will be kept in the permanent archive", Reg: rPlain, Form: fNotice, Topic: "authority"},
	{Text: "The last message sent out went to every address the city had ever held, including the ones dead for years, and it said thank you and goodbye in every language anybody there had ever spoken", Reg: rPlain, Topic: "message"},
}

// runEndSpace is the Space Age: the yards, the ring, the far side of the moon.
var runEndSpace = []skel{
	{Text: "The orbital yards were shut down", Reg: rPlain, Topic: "building"},
	{Text: "Launches were cancelled", Reg: rPlain, Topic: "machine"},
	{Text: "The last shuttle up carried seed stock, photographs and not much else", Reg: rPlain, Topic: "haul"},
	{Text: "Stations mothballed, crews rotated home, orbits logged", Reg: rPlain, Form: fLedger, Topic: "count"},
	{Text: "All craft are to return to dock and power down", Reg: rPlain, Form: fNotice, Topic: "authority"},
	{Text: "The moon base complains it was told last", Reg: rWry, Form: fComplaint, Topic: "authority"},
	{Text: "On the station they say the view never looked better", Reg: rWry, Form: fOverheard, Topic: "rumour"},
	{Text: "The mining crews came in from the belt", Reg: rPlain, Topic: "people"},
	{Text: "The ring station was turned so its windows faced home, and the crew took turns sitting at them through the last watch", Reg: rPlain, Topic: "omen"},
	{Text: "Up in orbit the satellites were left running on their own, and they went on sending down pictures of the weather over cities where the streets had emptied and the lights had been switched off one district at a time", Reg: rPlain, Topic: "machine"},
	{Text: "Mission control signed off", Reg: rPlain, Topic: "authority"},
	{Text: "The crew on the far side of the moon heard last of all, over a delayed signal, and wrote down every word of it by hand", Reg: rPlain, Topic: "message"},
	{Text: "Fuel went back into the tanks", Reg: rPlain, Topic: "kit"},
	{Text: "The orbital hotel refunded its bookings, less a handling fee", Reg: rJoke, Topic: "money"},
	{Text: "Children at the observation windows waved at the ships coming in until the last one had docked", Reg: rPlain, Topic: "family"},
}

// runEndCosmic is the Cosmic Era's shared voice: the age's harbinger at the
// end, and the dark it warned about. No speaker's own words; those live in
// runEndAges.
var runEndCosmic = []skel{
	{Text: "The final word went to {subject}", Needs: needSubject, Reg: rPlain, Topic: "message"},
	{Text: "Out past the last charted star something stirred, the way a sleeper stirs when a door closes in another part of the house", Reg: rPlain, Topic: "omen"},
	{Text: "Through the whole of the shutdown {subject} stayed, watching", Needs: needSubject, Reg: rPlain, Topic: "stranger"},
	{Text: "The dark came closer", Reg: rPlain, Topic: "omen"},
	{Text: "Whatever waits at the end of all roads had been told the colony was coming, and the last thing the long-range instruments recorded before they were powered down was a patch of sky where the background noise had fallen silent", Reg: rPlain, Topic: "omen"},
	{Text: "The crew complain that {subject} never said what comes after", Needs: needSubject, Reg: rWry, Form: fComplaint, Topic: "argument"},
	{Text: "All logs are to be left open for {subject} to read", Needs: needSubject, Reg: rWry, Form: fNotice, Topic: "authority"},
	{Text: "People on the command deck say {subject} knew the hour all along", Needs: needSubject, Reg: rWry, Form: fOverheard, Topic: "rumour"},
	{Text: "The stars went on", Reg: rPlain, Topic: "time"},
	{Text: "Air normal, power falling, every soul accounted for", Reg: rPlain, Form: fLedger, Topic: "count"},
	{Text: "When the colony's last lights went out, {subject} stayed on in the dark after the crew had gone to their berths", Needs: needSubject, Reg: rPlain, Topic: "stranger"},
	{Text: "The instruments kept recording after everyone had stopped reading them", Reg: rPlain, Topic: "machine"},
	{Text: "The colony ended its watch", Reg: rPlain, Topic: "work"},
	{Text: "Right up to the end {subject} offered no comfort and no blame", Needs: needSubject, Reg: rPlain, Topic: "stranger"},
	{Text: "Far out past the colony's instruments, where only one probe still reported, the thing it had been warned about for four ages turned away without hurry, and the probe lost track of it within the hour", Reg: rPlain, Topic: "omen"},
	{Text: "Engineering filed its last maintenance request, marked urgent", Reg: rJoke, Form: fNotice, Topic: "machine"},
	{Text: "The colonists took it in turns to sit with {subject} through the last watch, and none of them could say afterwards what had been said", Needs: needSubject, Reg: rPlain, Topic: "people"},
	{Text: "The children slept through it", Reg: rPlain, Topic: "family"},
	{Text: "The hydroponics bays were left lit", Reg: rPlain, Topic: "food"},
	{Text: "At the end the only voice left on the comms was {subject}", Needs: needSubject, Reg: rPlain, Topic: "noise"},
	{Text: "The old hands say the dark had been patient long enough, and that it was only ever waiting for the colony to stop", Reg: rPlain, Form: fOverheard, Topic: "rumour"},
}

// runEndAges is each Cosmic Era harbinger's own closing words, echoing its
// roster entry.
var runEndAges = []ageSet{
	{ages("interstellar_age"), []skel{
		{Text: "The beacon's loop ended in your name", Reg: rPlain, Topic: "name"},
		{Text: "The beacon kept looping after the last lights went out", Reg: rPlain, Topic: "noise"},
		{Text: "For eighty years the beacon had warned of this, and at the end it simply added the colony to its list", Reg: rPlain, Topic: "message"},
		{Text: "The dead colony's beacon and the living colony's last transmission overlapped for a few seconds, two recordings of the end playing over each other on the one frequency, and then the living one stopped", Reg: rPlain, Topic: "noise"},
		{Text: "The comms officer swears the beacon said goodbye", Reg: rWry, Form: fOverheard, Topic: "rumour"},
		{Text: "Beacon still transmitting, colony silent", Reg: rPlain, Form: fLedger, Topic: "count"},
	}},
	{ages("galactic_age"), []skel{
		{Text: "The relay folded the colony into its geometry and fell silent", Reg: rPlain, Topic: "omen"},
		{Text: "The linguists stayed at their posts to translate the last of it", Reg: rPlain, Topic: "work"},
		{Text: "The relay's hum held one note", Reg: rPlain, Topic: "noise"},
		{Text: "The relay had said this to others before, and the linguists found the colony's own name already written into its oldest geometry, among the names of peoples who had ended long before anyone here had learned to speak", Reg: rPlain, Topic: "name"},
		{Text: "The linguists complain that the relay's last message has no verb", Reg: rJoke, Form: fComplaint, Topic: "argument"},
		{Text: "Keep the relay powered after we are gone", Reg: rPlain, Form: fNotice, Topic: "authority"},
	}},
	{ages("quantum_age"), []skel{
		{Text: "Your future self sent nothing this time", Reg: rPlain, Topic: "message"},
		{Text: "From nine years out came one last message, and it was only your own breathing", Reg: rPlain, Topic: "noise"},
		{Text: "Your future self said this was the part you always forget, and that you should try to remember it this time", Reg: rWry, Topic: "time"},
		{Text: "The timestamps on your future self's messages had been running backwards for weeks, closer to today each time, and at the end they stopped there", Reg: rPlain, Topic: "time"},
		{Text: "Your future self said it had been worth it, which you chose to believe", Reg: rWry, Form: fOverheard, Topic: "rumour"},
		{Text: "In the last hour a message came from nine years out in your handwriting, the letters a little wrong, the hand that made them out of practice, and it said only that it was glad you had waited up", Reg: rPlain, Topic: "message"},
	}},
	{ages("transcendent_age"), []skel{
		{Text: "Your unmade self took your hand", Reg: rPlain, Topic: "stranger"},
		{Text: "This is how mine ended too, it said, and it did not let go", Reg: rPlain, Form: fOverheard, Topic: "stranger"},
		{Text: "The branch you stood on and the branch that ended met at last", Reg: rPlain, Topic: "time"},
		{Text: "Your unmade self had come all this way to see whether this branch ends the way its own did, and stayed to see that it did, and seemed almost relieved to have company", Reg: rWry, Topic: "stranger"},
		{Text: "Half the station complains that the other you never explains itself", Reg: rWry, Form: fComplaint, Topic: "argument"},
		{Text: "Every door on the station was unlocked, by order, so the other you could walk where it liked", Reg: rPlain, Form: fNotice, Topic: "authority"},
	}},
}
