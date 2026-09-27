package flavor

// HarbingerAppeased — the player spent faith and culture to lower the odds. The
// log line above says the odds went down; these sentences are what the
// appeasing looked like: what was given, who sang, who paid, and who is still
// complaining about the cost. Nothing here says whether it worked. That is the
// next Moment's job.
//
// The per-age pools echo config.HarbingerDef.AppeaseLabel for that age, so a
// player who clicked "Pay the monks to pray" reads about monks.

func harbAppeasedTemplates() []tmpl {
	out := pool("harb_appeased_any", erasAny, harbAppeasedAny)
	out = append(out, pool("harb_appeased_grounded", erasGrounded, harbAppeasedGrounded)...)
	out = append(out, pool("harb_appeased_late", erasLate, harbAppeasedLate)...)
	out = append(out, pool("harb_appeased_ancient", erasAncient, harbAppeasedAncient)...)
	out = append(out, pool("harb_appeased_feudal", erasFeudal, harbAppeasedFeudal)...)
	out = append(out, pool("harb_appeased_industrial", erasIndustrial, harbAppeasedIndustrial)...)
	out = append(out, pool("harb_appeased_digital", erasDigital, harbAppeasedDigital)...)
	out = append(out, pool("harb_appeased_cosmic", erasCosmic, harbAppeasedCosmic)...)
	out = append(out, agePools("harb_appeased_age", harbAppeasedAges)...)
	return out
}

var harbAppeasedAny = []skel{
	{Text: "Everyone gave something", Reg: rPlain, Topic: "haul"},
	{Text: "It took most of a day", Reg: rPlain, Topic: "time"},
	{Text: "The children were made to join in", Reg: rWry, Topic: "family"},
	{Text: "People who never pray prayed, awkwardly, standing at the back with their hands folded the way they had seen other people fold them", Reg: rWry, Topic: "religion"},
	{Text: "A good deal of what was given had been meant for other things", Reg: rPlain, Topic: "haul"},
	{Text: "An old man gave the thing he had been saving for his own funeral", Reg: rPlain, Topic: "people"},
	{Text: "Through all of it {subject} made no sign either way", Needs: needSubject, Reg: rPlain, Topic: "stranger"},
	{Text: "Everybody sang, including the people who cannot", Reg: rJoke, Topic: "noise"},
	{Text: "The mood afterwards was lighter than anyone could quite justify", Reg: rWry, Topic: "people"},
	{Text: "Several families gave more than they could spare and will be feeling it for a season", Reg: rPlain, Topic: "family"},
	{Text: "Whether any of it reached whoever it was meant for, nobody can say, but people walked home straighter than they had walked there", Reg: rWry, Form: fOverheard, Topic: "religion"},
	{Text: "The grumbling afterwards was mostly about the cost", Reg: rWry, Form: fComplaint, Topic: "money"},
	{Text: "A list was kept of who gave what", Reg: rPlain, Form: fLedger, Topic: "count"},
	{Text: "Afterwards there was an argument about who had given most, which went on well into the night and ended only when it was pointed out that the person who had given most had gone to bed without joining in", Reg: rWry, Form: fOverheard, Topic: "argument"},
}

var harbAppeasedGrounded = []skel{
	{Text: "Candles burned down to the holders", Reg: rPlain, Topic: "religion"},
	{Text: "Flowers were left on doorsteps", Reg: rPlain, Topic: "town"},
	{Text: "The dogs were shut in so they would not bark through it, and barked through it anyway from behind the doors", Reg: rPlain, Topic: "animal"},
	{Text: "Smoke from the offering fire hung over the roofs until evening", Reg: rPlain, Topic: "smell"},
	{Text: "A bench outside the shrine was worn smooth by the end of it", Reg: rWry, Topic: "building"},
	{Text: "The lamps were kept lit all night as part of it", Reg: rPlain, Topic: "sleep"},
	{Text: "The women who organised it are already talking about doing it every year, and the men who carried the heavy things are already talking about being elsewhere next year", Reg: rJoke, Form: fOverheard, Topic: "argument"},
	{Text: "At the end of it the oldest woman in the place stood up, walked three times round the fire with her eyes shut, sat down again, and said that would do, and everyone went home as if a door had closed", Reg: rPlain, Topic: "people"},
}

var harbAppeasedLate = []skel{
	{Text: "Everyone tuned in", Reg: rPlain, Topic: "people"},
	{Text: "The screens were switched off for an hour", Reg: rPlain, Topic: "machine"},
	{Text: "The whole thing was streamed, and most people watched it alone in their rooms and then went out into the corridors afterwards to be near somebody", Reg: rPlain, Topic: "machine"},
	{Text: "Donations came in faster than the servers could log them", Reg: rPlain, Form: fLedger, Topic: "money"},
	{Text: "The hydroponics crew gave their first crop of the year", Reg: rPlain, Topic: "food"},
	{Text: "A minute of silence was held on every deck and every screen", Reg: rPlain, Form: fNotice, Topic: "noise"},
	{Text: "The ceremony was put together from half-remembered pieces of older ones, because nobody alive had done this properly, and it was moving anyway", Reg: rWry, Topic: "religion"},
	{Text: "For an hour the settlement's systems ran nothing but the ceremony, every display and speaker given over to it, and afterwards people said the silence when the machines went back to their ordinary work had felt like being set down", Reg: rPlain, Topic: "machine"},
}

var harbAppeasedAncient = []skel{
	{Text: "A goat went to the stones", Reg: rPlain, Topic: "animal"},
	{Text: "The elders danced", Reg: rPlain, Topic: "authority"},
	{Text: "The drums went all night, slow", Reg: rPlain, Topic: "noise"},
	{Text: "Bundles of sweet grass were burned at every hut until the whole camp smelled of it, and the smell was still in people's hair three days later", Reg: rPlain, Topic: "smell"},
	{Text: "The best of the dried meat went into the fire, and the complaining was kept low", Reg: rWry, Topic: "food"},
	{Text: "The children were painted with ochre and made to stand still while the elders sang", Reg: rPlain, Topic: "family"},
	{Text: "The offerings were carried up to the standing stones in a line, oldest first", Reg: rPlain, Topic: "religion"},
	{Text: "At the stones they left ~ and walked away without looking back", Slot: "harb_offering_ancient", Reg: rPlain, Topic: "haul"},
	{Text: "Each family brought ~ and laid it at the foot of the tallest stone", Slot: "harb_offering_ancient", Reg: rPlain, Topic: "religion"},
	{Text: "The hunters gave up their best spear points, which they will regret the next time a herd comes through, and they know it", Reg: rWry, Topic: "weapon"},
	{Text: "The elders say it was done properly", Reg: rPlain, Form: fOverheard, Topic: "authority"},
	{Text: "Two of the hunters grumble that the goat was theirs", Reg: rWry, Form: fComplaint, Topic: "animal"},
	{Text: "The elders have said no meat is to be eaten until the moon is full", Reg: rPlain, Form: fNotice, Topic: "food"},
	{Text: "They built the fire higher than anyone could remember seeing it and fed it everything they had agreed to give, and then some things they had not agreed to give, and stood round it until it burned down to a red eye", Reg: rPlain, Topic: "haul"},
}

var harbAppeasedFeudal = []skel{
	{Text: "The bells rang for an hour", Reg: rPlain, Topic: "noise"},
	{Text: "Masses were said for three days", Reg: rPlain, Topic: "religion"},
	{Text: "The abbot sent a bill", Reg: rJoke, Form: fLedger, Topic: "money"},
	{Text: "The pardoner did well", Reg: rWry, Topic: "trade"},
	{Text: "The relics were carried round the walls under a canopy, and the canopy caught on the gatehouse and had to be taken down and put up again on the far side", Reg: rPlain, Topic: "religion"},
	{Text: "Every household gave ~ to the church", Slot: "harb_offering_feudal", Reg: rPlain, Topic: "haul"},
	{Text: "Among the offerings was ~, which will be talked about for years", Slot: "harb_offering_feudal", Reg: rWry, Topic: "rumour"},
	{Text: "A procession went out at dawn with the bishop at the front and most of the parish behind", Reg: rPlain, Topic: "town"},
	{Text: "Pilgrims were sent to the shrine with a purse and a list of names", Reg: rPlain, Topic: "name"},
	{Text: "The monks prayed in turns through the night, and the abbey kitchen fed them in turns too", Reg: rPlain, Topic: "food"},
	{Text: "Alms were given at the church door to anyone who came", Reg: rPlain, Topic: "money"},
	{Text: "The steward grumbles that the lord's wine went to the altar", Reg: rWry, Form: fComplaint, Topic: "authority"},
	{Text: "It was proclaimed at the market cross that there would be fasting on Friday", Reg: rPlain, Form: fNotice, Topic: "food"},
	{Text: "The guilds each paid for a candle as tall as a man and tried to make theirs taller than the others', and by evening the cathedral was so bright with guild candles that the priest could not see the altar", Reg: rJoke, Topic: "argument"},
}

var harbAppeasedIndustrial = []skel{
	{Text: "The collection plates came back full", Reg: rPlain, Topic: "money"},
	{Text: "A day of prayer was declared", Reg: rPlain, Form: fNotice, Topic: "religion"},
	{Text: "The mayor gave generously, and publicly", Reg: rWry, Topic: "authority"},
	{Text: "The brass band played in the park until dark", Reg: rPlain, Topic: "noise"},
	{Text: "Factory owners gave a day's takings to the churches, and had their names read out for it", Reg: rWry, Topic: "name"},
	{Text: "The choirs of four chapels sang together, which they had never managed before", Reg: rPlain, Topic: "religion"},
	{Text: "Subscriptions were taken up in every pub in the district, and the landlords matched them, and then put their prices up a farthing the following week", Reg: rPlain, Topic: "money"},
	{Text: "The newspapers printed the names of everyone who gave", Reg: rPlain, Topic: "paper"},
	{Text: "The revival tent on the common was full every night", Reg: rPlain, Topic: "town"},
	{Text: "The museum lent its oldest relic for the service at the cathedral, and a queue formed down the length of the high street to walk past it", Reg: rPlain, Topic: "building"},
	{Text: "Down at the works the men took up a collection on the shop floor, a penny or two each, and the foreman put in a shilling and made sure everyone saw, and the whole of it went to the chapel in a biscuit tin", Reg: rWry, Topic: "work"},
	{Text: "The treasurer complains that half the pledges are still unpaid", Reg: rWry, Form: fComplaint, Topic: "money"},
	{Text: "Four thousand pledged, two thousand collected, and a great deal of goodwill", Reg: rWry, Form: fLedger, Topic: "count"},
}

var harbAppeasedDigital = []skel{
	{Text: "The hashtag trended for two days", Reg: rPlain, Topic: "rumour"},
	{Text: "Donations were matched by three corporations", Reg: rPlain, Topic: "money"},
	{Text: "Tribute pages went up everywhere", Reg: rPlain, Topic: "message"},
	{Text: "A billion in pledges by midnight", Reg: rPlain, Form: fLedger, Topic: "count"},
	{Text: "An online vigil drew more viewers than the final of anything", Reg: rWry, Topic: "people"},
	{Text: "The meditation apps offered a free month, and people took it", Reg: rPlain, Topic: "machine"},
	{Text: "The feed was full of people asking to be forgiven for things", Reg: rPlain, Topic: "religion"},
	{Text: "The city's arts fund paid for a week of free concerts, and the orchestra played to half-empty halls because everyone was watching at home", Reg: rWry, Topic: "noise"},
	{Text: "Corporate sponsors had their logos on the prayer broadcast", Reg: rJoke, Topic: "money"},
	{Text: "Churches, temples and mosques shared a single stream for the first time, and the comments underneath were kinder than anyone expected them to be", Reg: rPlain, Topic: "religion"},
	{Text: "In the server halls the engineers held a vigil of their own, standing in the cold aisles between the racks with the lights off and the fans roaring, and at least one of them said afterwards that it had helped", Reg: rPlain, Topic: "machine"},
	{Text: "The comments complain it was all for show", Reg: rWry, Form: fComplaint, Topic: "argument"},
	{Text: "The public broadcaster suspended its schedule for the ceremony", Reg: rPlain, Form: fNotice, Topic: "authority"},
}

var harbAppeasedCosmic = []skel{
	{Text: "The station went dark", Reg: rPlain, Topic: "building"},
	{Text: "The ring held its breath", Reg: rPlain, Topic: "noise"},
	{Text: "A hymn was broadcast into the dark", Reg: rPlain, Topic: "religion"},
	{Text: "The observation deck was filled with flowers from hydroponics", Reg: rPlain, Topic: "food"},
	{Text: "Every hull light was dimmed for an hour", Reg: rPlain, Topic: "machine"},
	{Text: "The oldest recordings in the archive were played in full", Reg: rPlain, Topic: "noise"},
	{Text: "Everyone wrote a message on a slip of film, and the slips were sealed in a capsule and fired into space at a slow drift, to be found or not", Reg: rPlain, Topic: "message"},
	{Text: "The fleet flew in formation past the station's windows", Reg: rPlain, Topic: "machine"},
	{Text: "The chaplains of four faiths stood together on the hangar floor and led a service that was partly each of theirs and wholly none of them", Reg: rWry, Topic: "religion"},
	{Text: "Out beyond the last marker buoy they set adrift a small capsule carrying the names of everyone in the colony and a recording of the children singing, on a course that will take it nowhere for longer than the colony will last", Reg: rPlain, Topic: "name"},
	{Text: "Engineering complains that the power for the ceremony came out of their allocation", Reg: rWry, Form: fComplaint, Topic: "machine"},
	{Text: "All nonessential systems were powered down for the vigil by order of the commander", Reg: rPlain, Form: fNotice, Topic: "authority"},
	{Text: "Flowers from hydroponics, one crate per deck", Reg: rPlain, Form: fLedger, Topic: "count"},
}

// harbAppeasedAges echoes each age's AppeaseLabel.
var harbAppeasedAges = []ageSet{
	{ages("primitive_age"), []skel{
		{Text: "The offerings at the stones were piled higher than a child", Reg: rPlain, Topic: "haul"},
		{Text: "By the evening there was so much left at the stones that the foxes came, and the elders said the foxes were a good sign", Reg: rWry, Topic: "animal"},
	}},
	{ages("stone_age"), []skel{
		{Text: "His sign went up on the cave wall, cut deep", Reg: rPlain, Topic: "building"},
		{Text: "Every hand cut a stroke", Reg: rPlain, Topic: "people"},
	}},
	{ages("bronze_age"), []skel{
		{Text: "The goat went to the altar without a fuss", Reg: rPlain, Topic: "animal"},
		{Text: "The priest who cut the goat's throat had done it a hundred times before and still took a long breath first, and the smoke from the altar went straight up into a sky without a breath of wind in it, which everybody agreed was right", Reg: rPlain, Topic: "religion"},
	}},
	{ages("iron_age"), []skel{
		{Text: "The sackcloth itched terribly", Reg: rWry, Form: fComplaint, Topic: "kit"},
		{Text: "The whole city fasted for three days, even the tax collectors", Reg: rWry, Topic: "food"},
	}},
	{ages("classical_age"), []skel{
		{Text: "The gifts went to the shrine on nine mules", Reg: rPlain, Topic: "haul"},
		{Text: "The priests of the shrine received the gifts, counted them carefully, and sent back word that the god was paying attention", Reg: rWry, Topic: "religion"},
	}},
	{ages("medieval_age"), []skel{
		{Text: "The monks were paid, and prayed", Reg: rPlain, Topic: "religion"},
		{Text: "The abbey's prayers went on day and night in shifts of three monks, and the town paid for every hour of it", Reg: rWry, Topic: "money"},
	}},
	{ages("renaissance_age"), []skel{
		{Text: "The painter was paid half in advance", Reg: rPlain, Form: fLedger, Topic: "money"},
		{Text: "The altarpiece shows the city kneeling under a stormy sky, and the prince has had himself painted at the front, larger than the saints", Reg: rWry, Topic: "authority"},
	}},
	{ages("colonial_age"), []skel{
		{Text: "The governor proclaimed the fast from the courthouse steps", Reg: rPlain, Form: fNotice, Topic: "authority"},
		{Text: "Taverns shut for the day", Reg: rPlain, Topic: "trade"},
	}},
	{ages("industrial_age"), []skel{
		{Text: "The revival preacher came up from the coast by train", Reg: rPlain, Topic: "stranger"},
		{Text: "The revival tent held two thousand and there were people standing outside it in the rain, singing along through the canvas", Reg: rPlain, Topic: "religion"},
	}},
	{ages("victorian_age"), []skel{
		{Text: "The national day of prayer closed the mills", Reg: rPlain, Topic: "work"},
		{Text: "Every church read the proclamation", Reg: rPlain, Form: fNotice, Topic: "religion"},
	}},
	{ages("electric_age"), []skel{
		{Text: "The revival tour went out by special train with electric lights on every carriage", Reg: rPlain, Topic: "machine"},
		{Text: "The tour's preacher spoke from a stage lit by more electric bulbs than the town had ever seen in one place, and people came as much for the bulbs", Reg: rJoke, Topic: "people"},
	}},
	{ages("atomic_age"), []skel{
		{Text: "The new early-warning stations went up along the coast, paid for by public subscription", Reg: rPlain, Topic: "building"},
		{Text: "Radar dishes turned on every headland", Reg: rPlain, Topic: "machine"},
	}},
	{ages("modern_age"), []skel{
		{Text: "The telethon ran for thirty hours", Reg: rPlain, Topic: "time"},
		{Text: "Celebrities answered the phones on the telethon, badly, and the total on the board behind them went up all night", Reg: rWry, Topic: "money"},
	}},
	{ages("information_age"), []skel{
		{Text: "Everybody forwarded it to everybody they knew", Reg: rPlain, Topic: "message"},
		{Text: "It went to every address in every address book, and to people who had not heard from the sender in fifteen years", Reg: rWry, Topic: "people"},
	}},
	{ages("digital_age"), []skel{
		{Text: "The feel-good hashtag was everywhere by lunch", Reg: rPlain, Topic: "rumour"},
		{Text: "Influencers were paid to post calm", Reg: rWry, Topic: "money"},
	}},
	{ages("cyberpunk_age"), []skel{
		{Text: "The oracle AI took the bribe and said nothing about it", Reg: rWry, Topic: "money"},
		{Text: "The bribe went through a chain of shell accounts so long that the oracle's own auditors lost track of it halfway", Reg: rWry, Topic: "money"},
	}},
	{ages("fusion_age"), []skel{
		{Text: "The vigil in the containment hall went on through two shifts", Reg: rPlain, Topic: "religion"},
		{Text: "Engineers left flowers on the reactor housing", Reg: rPlain, Topic: "machine"},
	}},
	{ages("space_age"), []skel{
		{Text: "The peace hymn went out on every frequency", Reg: rPlain, Topic: "noise"},
		{Text: "The hymn was sung by the station's children and sent outward at full power, and it will still be travelling long after anyone who sang it is gone", Reg: rPlain, Topic: "family"},
	}},
	{ages("interstellar_age"), []skel{
		{Text: "The names of the lost colonists were read aloud", Reg: rPlain, Topic: "name"},
		{Text: "The lost colony's anthem was played on every deck", Reg: rPlain, Topic: "noise"},
	}},
	{ages("galactic_age"), []skel{
		{Text: "The linguists answered the relay in its own geometry", Reg: rPlain, Topic: "message"},
		{Text: "It took the linguists four days to compose the reply, and they argued over every angle of it, and when it was sent the relay hummed once and fell silent", Reg: rPlain, Topic: "argument"},
	}},
	{ages("quantum_age"), []skel{
		{Text: "You kept the promises, even the small ones", Reg: rPlain, Topic: "time"},
		{Text: "You did the things your future self had requested, one by one, including the one you had been putting off for years, and it was harder than it should have been", Reg: rWry, Topic: "work"},
	}},
	{ages("transcendent_age"), []skel{
		{Text: "The branches that ended were mourned by name", Reg: rPlain, Topic: "name"},
		{Text: "You held a service for the versions of yourself that did not continue, and the hall was full, and the strange thing was that nobody present was sure how many of the mourners had come from this branch at all", Reg: rPlain, Topic: "religion"},
	}},
}
