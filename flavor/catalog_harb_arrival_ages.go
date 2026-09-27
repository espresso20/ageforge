package flavor

// harbArrivalAges is the figure itself, one batch per speaker (see
// config.HarbingerFor for the roster these are written against).
var harbArrivalAges = []ageSet{
	{ages("primitive_age"), []skel{
		{Text: "He smelled of smoke and carrion", Reg: rPlain, Topic: "smell"},
		{Text: "He came out of the trees on the far side of the old burn, grey with ash from head to foot", Reg: rPlain, Topic: "stranger"},
		{Text: "A string of finger bones hung round his neck, and he would not say whose", Reg: rWry, Topic: "stranger"},
		{Text: "He squatted by the fire pit and drew circles in the ash with one long black nail, one inside the other, until the whole camp had gathered round to watch and the smallest children had been lifted up to see over the rest", Reg: rPlain, Topic: "omen"},
	}},
	{ages("stone_age"), []skel{
		{Text: "He had walked for days", Reg: rPlain, Topic: "ground"},
		{Text: "Three knots in his beard, one for each thing he had seen", Reg: rWry, Topic: "omen"},
		{Text: "He came down from the high caves for the first time since the long winter the old women still talk about", Reg: rPlain, Topic: "stranger"},
		{Text: "Meat was brought to him and refused, and he sat at the cave mouth until dark", Reg: rPlain, Topic: "food"},
	}},
	{ages("bronze_age"), []skel{
		{Text: "She came on a mule with a bag of knucklebones, a cage of sparrows and a boy to carry both", Reg: rPlain, Topic: "stranger"},
		{Text: "First, she wanted a goat", Reg: rWry, Form: fComplaint, Topic: "animal"},
		{Text: "Her hands were stained brown to the wrist from reading entrails", Reg: rPlain, Topic: "omen"},
		{Text: "The temple priests would not meet her eye", Reg: rWry, Topic: "religion"},
	}},
	{ages("iron_age"), []skel{
		{Text: "Sand in his beard, and nothing on his feet", Reg: rPlain, Topic: "stranger"},
		{Text: "He had eaten locusts on the road", Reg: rJoke, Topic: "food"},
		{Text: "He walked in from the dry country in a camel-hair cloak, stood by the well at noon, and shouted until the whole market came to look", Reg: rPlain, Topic: "noise"},
		{Text: "The dogs trailed him in from the edge of the camp", Reg: rPlain, Topic: "animal"},
	}},
	{ages("classical_age"), []skel{
		{Text: "The pilgrims came back grey", Reg: rPlain, Topic: "religion"},
		{Text: "A priest of the shrine rode in with a sealed tablet and a sore back", Reg: rWry, Topic: "message"},
		{Text: "The smoke over the cleft rock had turned black, the priests said, and she had spoken without being asked a question, which in living memory she had never once done, and the priests had needed three days to agree on what she meant", Reg: rPlain, Form: fOverheard, Topic: "omen"},
		{Text: "The messenger wanted feeding, stabling and his fee before he would break a seal", Reg: rJoke, Form: fComplaint, Topic: "money"},
	}},
	{ages("medieval_age"), []skel{
		{Text: "He rang his bell at the market cross until his arm gave out", Reg: rPlain, Topic: "noise"},
		{Text: "Red coat, gone pink with washing", Reg: rPlain, Topic: "stranger"},
		{Text: "He climbed onto the well-head to be seen, and before he had read a word of it the people at the back were already crossing themselves", Reg: rPlain, Topic: "religion"},
		{Text: "His voice cracked on the second line and he started again from the top", Reg: rWry, Topic: "noise"},
	}},
	{ages("renaissance_age"), []skel{
		{Text: "He came down from the palace in a hired coach with the charts rolled under one arm and a velvet bag of instruments under the other", Reg: rPlain, Topic: "paper"},
		{Text: "His astrolabe caught the sun", Reg: rPlain, Topic: "machine"},
		{Text: "The astrologer asked for a table, a candle and silence, and got two of them", Reg: rJoke, Topic: "authority"},
		{Text: "The prince's horoscope had been cast three times, and he liked none of them", Reg: rWry, Topic: "omen"},
	}},
	{ages("colonial_age"), []skel{
		{Text: "She set up outside the coffee house with a satchel of pamphlets still wet from the press and a voice like a gull", Reg: rPlain, Topic: "paper"},
		{Text: "A penny a pamphlet", Reg: rPlain, Form: fLedger, Topic: "money"},
		{Text: "Her printer's apron was black to the elbow", Reg: rPlain, Topic: "stranger"},
		{Text: "The governor's men watched her from across the square with their arms folded, doing nothing yet, and she read the whole pamphlet out a second time for their benefit, slowly, as if they were hard of hearing", Reg: rWry, Topic: "authority"},
	}},
	{ages("industrial_age"), []skel{
		{Text: "The newsboy on the corner had a voice that carried three streets", Reg: rPlain, Topic: "noise"},
		{Text: "Cap too big, bag too heavy", Reg: rPlain, Topic: "stranger"},
		{Text: "The presses had run an extra edition at noon, and he was the first boy out of the yard with it, still folding as he ran", Reg: rPlain, Topic: "paper"},
		{Text: "He took a halfpenny off everyone who stopped to listen", Reg: rWry, Topic: "money"},
	}},
	{ages("victorian_age"), []skel{
		{Text: "He stood on a soap crate by the park railings with a sandwich board over his shoulders and his hat held out in one hand", Reg: rPlain, Topic: "stranger"},
		{Text: "THE END IS NIGH, hand-painted", Reg: rPlain, Topic: "message"},
		{Text: "A constable moved him along and he set up again twenty yards further down", Reg: rWry, Topic: "authority"},
		{Text: "A table of figures was pasted to the back of the board", Reg: rPlain, Topic: "paper"},
	}},
	{ages("electric_age"), []skel{
		{Text: "The telegraph at the post office started chattering just after midnight", Reg: rPlain, Topic: "machine"},
		{Text: "The operator came in slippers", Reg: rWry, Topic: "people"},
		{Text: "The dispatches came from stations up the line one after another, and the operator took each one down in pencil, and his hand shook enough by the fourth that he had to write it out again for the postmaster to read", Reg: rPlain, Topic: "paper"},
		{Text: "The postmaster sat beside the key with his watch open", Reg: rPlain, Topic: "time"},
	}},
	{ages("atomic_age"), []skel{
		{Text: "The programme stopped mid-song", Reg: rPlain, Topic: "noise"},
		{Text: "The civil defence tone came over the wireless, three long notes and a pause, and then the announcer, reading from a card", Reg: rPlain, Topic: "noise"},
		{Text: "Every radio in the street was turned up", Reg: rPlain, Topic: "machine"},
		{Text: "The announcer had the calm, level voice they train them to have", Reg: rWry, Topic: "people"},
	}},
	{ages("modern_age"), []skel{
		{Text: "The evening news led with it", Reg: rPlain, Topic: "message"},
		{Text: "The anchor came on ten minutes early without a tie", Reg: rWry, Topic: "people"},
		{Text: "A red strip ran along the bottom of the screen all evening, and in house after house people sat through the adverts rather than turn over", Reg: rPlain, Topic: "machine"},
		{Text: "The weather segment was dropped, which people noticed", Reg: rWry, Topic: "weather"},
	}},
	{ages("information_age"), []skel{
		{Text: "It arrived in forty inboxes before breakfast, subject line in capitals", Reg: rPlain, Topic: "message"},
		{Text: "The email had been forwarded so many times that the original sender was buried under a hundred signatures and a great many holiday auto-replies", Reg: rWry, Topic: "message"},
		{Text: "The attachment went unopened", Reg: rPlain, Topic: "paper"},
		{Text: "Everybody's aunt had sent it to everybody, with a note saying she did not usually forward these", Reg: rJoke, Topic: "family"},
	}},
	{ages("digital_age"), []skel{
		{Text: "A million views by lunch", Reg: rPlain, Form: fLedger, Topic: "count"},
		{Text: "It was shot on a phone, shaky, in portrait", Reg: rPlain, Topic: "machine"},
		{Text: "The thumbnail alone was enough to stop people scrolling, and the first few seconds were enough to make them turn the sound up", Reg: rPlain, Topic: "people"},
		{Text: "Reaction videos to the video were up within the hour", Reg: rWry, Topic: "rumour"},
	}},
	{ages("cyberpunk_age"), []skel{
		{Text: "It surfaced in the corporate net at three minutes past midnight, speaking through the account of an executive who had been dead since the last merger", Reg: rPlain, Topic: "machine"},
		{Text: "Every screen flickered once", Reg: rPlain, Topic: "machine"},
		{Text: "Corporate security tried to purge it and failed", Reg: rPlain, Topic: "authority"},
		{Text: "Its voice was stitched together from old customer-service recordings", Reg: rPlain, Topic: "noise"},
	}},
	{ages("fusion_age"), []skel{
		{Text: "It spoke over every speaker in the plant at once, in the level, pleasant voice it normally keeps for fire drills and visitors", Reg: rPlain, Topic: "noise"},
		{Text: "Containment lights went amber", Reg: rPlain, Topic: "machine"},
		{Text: "The warden had never spoken outside a scheduled drill before", Reg: rPlain, Topic: "machine"},
		{Text: "Engineers came in on their day off without being called", Reg: rPlain, Topic: "work"},
	}},
	{ages("space_age"), []skel{
		{Text: "The monitoring station behind the moon sent a priority packet", Reg: rPlain, Topic: "message"},
		{Text: "Its sensors had been pointed at one patch of sky for forty years without finding anything worth waking a person for, and the packet it sent was the first it had ever marked urgent, and it had marked every line of it", Reg: rPlain, Topic: "machine"},
		{Text: "The duty officer read it aloud", Reg: rPlain, Topic: "people"},
		{Text: "The station's automatic voice was flat and polite", Reg: rPlain, Topic: "noise"},
	}},
	{ages("interstellar_age"), []skel{
		{Text: "The beacon came from a colony that went silent eighty years ago", Reg: rPlain, Topic: "message"},
		{Text: "It is still transmitting on the old emergency band, on a loop, in a voice everybody's grandparents would have known", Reg: rPlain, Topic: "noise"},
		{Text: "The colony is listed as lost", Reg: rPlain, Form: fLedger, Topic: "name"},
		{Text: "Nobody has ever gone out there to switch it off", Reg: rWry, Topic: "time"},
	}},
	{ages("galactic_age"), []skel{
		{Text: "The relay predates us", Reg: rPlain, Topic: "machine"},
		{Text: "It woke without warning and began to hum on a frequency the instruments had to be rebuilt to hear", Reg: rPlain, Topic: "noise"},
		{Text: "Its message arrived as geometry, and the linguists needed a day to make it into words", Reg: rPlain, Topic: "message"},
		{Text: "The relay last spoke when there were no people anywhere", Reg: rPlain, Topic: "time"},
	}},
	{ages("quantum_age"), []skel{
		{Text: "The message arrived stamped with a date nine years from now", Reg: rPlain, Topic: "time"},
		{Text: "Your name, your handwriting", Reg: rPlain, Topic: "name"},
		{Text: "The voice was yours, older and hoarse", Reg: rPlain, Topic: "noise"},
		{Text: "Whoever sent it knew the passcode only you know, and the name you had for yourself as a child", Reg: rPlain, Topic: "message"},
	}},
	{ages("transcendent_age"), []skel{
		{Text: "From a branch that ended", Reg: rPlain, Topic: "omen"},
		{Text: "It had your face, or had once", Reg: rPlain, Topic: "stranger"},
		{Text: "It looked out of every reflective surface in the settlement at once, for about the length of a breath", Reg: rPlain, Topic: "omen"},
		{Text: "There was no moment of arrival, only a slow sense that it had been standing among you for some time, and that it had been waiting, with more patience than you have ever had, for you to notice", Reg: rPlain, Topic: "time"},
	}},
}
