package mapgen

// templates are the wonders drawn so far, by wonder key, each small and drawn
// by hand from its name and what the game says of it. The engine sets one on
// its plot and dresses it in the age's materials. A wonder without one is
// drawn as a generic monument (wonder.go).
var templates = map[string]Template{
	"sacred_grove": {
		Art: []string{
			"  ♣ ♣ ♣  ",
			" ♣     ♣ ",
			"♣   ∩   ♣",
			" ♣     ♣ ",
			"  ♣ ♣ ♣  ",
		},
		Tone: []string{
			"  v v v  ",
			" v     v ",
			"v   s   v",
			" v     v ",
			"  v v v  ",
		},
	},
	"great_monolith": {
		Art: []string{
			"   ▲   ",
			"  ▐█▌  ",
			"  ▐█▌  ",
			"  ▐█▌  ",
			" ▄▄█▄▄ ",
		},
		Tone: []string{
			"   a   ",
			"  sss  ",
			"  sss  ",
			"  sss  ",
			" sssss ",
		},
	},
	"colosseum": {
		Art: []string{
			" ▄▀▀▀▀▀▀▄ ",
			"▟▌◠◠◠◠◠◠▐▙",
			"▜▌◠◠◠◠◠◠▐▛",
			" ▀▄▄▄▄▄▄▀ ",
		},
		Tone: []string{
			" ssssssss ",
			"ssmmmmmmss",
			"ssmmmmmmss",
			" ssssssss ",
		},
	},
	"parthenon": {
		Art: []string{
			" ▄▄▄▄▄▄▄ ",
			"▐███████▌",
			" ║║║║║║║ ",
			" ║║║║║║║ ",
			"▀▀▀▀▀▀▀▀▀",
		},
	},
	"great_library": {
		Art: []string{
			"   ▄◠▄   ",
			" ▄█████▄ ",
			" █▫▫▫▫▫█ ",
			" █▫▫∩▫▫█ ",
			" ▀▀▀▀▀▀▀ ",
		},
		Tone: []string{
			"   sas   ",
			" sssssss ",
			" sgggggs ",
			" sggsggs ",
			" sssssss ",
		},
	},
	"sistine_chapel": {
		Art: []string{
			"    †    ",
			"   ▄█▄   ",
			" ▄▟███▙▄ ",
			" █▫▫∩▫▫█ ",
			" ▀▀▀▀▀▀▀ ",
		},
		Tone: []string{
			"    a    ",
			"   sss   ",
			" sssssss ",
			" sggsggs ",
			" sssssss ",
		},
	},
	"grand_lighthouse": {
		Art: []string{
			"   ─☼─   ",
			"    █    ",
			"   ▐█▌   ",
			"   ▐█▌   ",
			"  ▄███▄  ",
		},
		Tone: []string{
			"   fff   ",
			"    s    ",
			"   sss   ",
			"   sss   ",
			"  sssss  ",
		},
	},
	"crystal_palace": {
		Art: []string{
			"  ▄▄▄▄▄  ",
			" ◢▒▒▒▒▒◣ ",
			"◢▒▒▒▒▒▒▒◣",
			"█▒▒▒▒▒▒▒█",
			"▀▀▀▀▀▀▀▀▀",
		},
		Tone: []string{
			"  mmmmm  ",
			" mgggggm ",
			"mgggggggm",
			"mgggggggm",
			"mmmmmmmmm",
		},
	},
	"eiffel_tower": {
		Art: []string{
			"    ▲    ",
			"   ╱╳╲   ",
			"  ╱╫╫╫╲  ",
			" ╱╱ ╳ ╲╲ ",
			"╱╱  ╳  ╲╲",
		},
		Tone: []string{
			"    a    ",
			"   mmm   ",
			"  mmmmm  ",
			" mm m mm ",
			"mm  m  mm",
		},
	},
	"hoover_dam": {
		Art: []string{
			"≈≈≈≈▐▀▀▌  ",
			"≈≈≈≈▐██▌▒▒",
			"≈≈≈▟███▌▒▒",
			"≈≈▟████▌▒·",
			"≈▟█████▌▒ ",
		},
		Tone: []string{
			"wwwwssss  ",
			"wwwwssssww",
			"wwwsssssww",
			"wwssssssww",
			"wssssssssw",
		},
	},
	"particle_accelerator": {
		Art: []string{
			" ╭──────╮ ",
			"╭╯╭────╮╰╮",
			"│ │ ◘◘ │ │",
			"╰╮╰────╯╭╯",
			" ╰──────╯ ",
		},
		Tone: []string{
			" mmmmmmmm ",
			"mmmmmmmmmm",
			"m m ff m m",
			"mmmmmmmmmm",
			" mmmmmmmm ",
		},
	},
	"global_network": {
		Art: []string{
			" ◢◣ ╽ ◢◣ ",
			" ╲╱ █ ╲╱ ",
			"  ╲ █ ╱  ",
			"▄▄▄▟█▙▄▄▄",
			"▀▀▀▀▀▀▀▀▀",
		},
		Tone: []string{
			" mm g mm ",
			" mm s mm ",
			"  m s m  ",
			"sssssssss",
			"sssssssss",
		},
	},
	"world_simulation": {
		Art: []string{
			"╔═══════╗",
			"║ ▒▒▒▒▒ ║",
			"║ ▒◘◘◘▒ ║",
			"║ ▒▒▒▒▒ ║",
			"╚═══════╝",
		},
		Tone: []string{
			"mmmmmmmmm",
			"m ggggg m",
			"m gnnng m",
			"m ggggg m",
			"mmmmmmmmm",
		},
	},
	"neon_citadel": {
		Art: []string{
			"▟▙  ▲  ▟▙",
			"█▪▫▪█▪▫▪█",
			"█▫▪▫▫▫▪▫█",
			"█▪▫▪∩▪▫▪█",
			"▀▀▀▀▀▀▀▀▀",
		},
		Tone: []string{
			"ss  a  ss",
			"sngnsngns",
			"sgngggngs",
			"sngnsngns",
			"sssssssss",
		},
	},
	"stellar_cradle": {
		Art: []string{
			"  ╭───╮  ",
			" ╭╯ ☼ ╰╮ ",
			" │ ☼☼☼ │ ",
			" ╰╮ ☼ ╭╯ ",
			"  ╰───╯  ",
		},
		Tone: []string{
			"  mmmmm  ",
			" mm f mm ",
			" m fff m ",
			" mm f mm ",
			"  mmmmm  ",
		},
	},
	"warp_nexus": {
		Art: []string{
			"  ╭───╮  ",
			" ╭╯ ▒ ╰╮ ",
			"─┤ ▒▒▒ ├─",
			" ╰╮ ▒ ╭╯ ",
			"  ╰───╯  ",
		},
		Tone: []string{
			"  mmmmm  ",
			" mm g mm ",
			"mm ggg mm",
			" mm g mm ",
			"  mmmmm  ",
		},
	},
	"cosmic_beacon": {
		Art: []string{
			"  ╲ ║ ╱  ",
			"   ╲║╱   ",
			"────☼────",
			"   ╱║╲   ",
			"  ╱ ║ ╲  ",
		},
		Tone: []string{
			"  a a a  ",
			"   aaa   ",
			"aaaafaaaa",
			"   aaa   ",
			"  a a a  ",
		},
	},
	"reality_anchor": {
		Art: []string{
			"   ╭─╮   ",
			"   ╰┬╯   ",
			"  ──┼──  ",
			" ╲  │  ╱ ",
			"  ╰═╧═╯  ",
		},
		Tone: []string{
			"   nnn   ",
			"   nmn   ",
			"  nnmnn  ",
			" n  m  n ",
			"  nnmnn  ",
		},
	},
	"singularity_core": {
		Art: []string{
			"  ╭▒▒▒╮  ",
			" ▒▒▓▓▓▒▒ ",
			"▒▓▓███▓▓▒",
			" ▒▒▓▓▓▒▒ ",
			"  ╰▒▒▒╯  ",
		},
		Tone: []string{
			"  nnnnn  ",
			" nnfffnn ",
			"nffsssffn",
			" nnfffnn ",
			"  nnnnn  ",
		},
	},
	"stonehenge": {
		Art: []string{
			"  Π Π Π  ",
			" Π     Π ",
			"Π   ▄   Π",
			" Π     Π ",
			"  Π Π Π  ",
		},
		Tone: []string{
			"  s s s  ",
			" s     s ",
			"s   a   s",
			" s     s ",
			"  s s s  ",
		},
	},
	"space_program": {
		Art: []string{
			"    ▲    ",
			"   ╱█╲   ",
			"  ╟│█│╢  ",
			" ╤═╧█╧═╤ ",
			"▄▄▄▄▄▄▄▄▄",
		},
		Tone: []string{
			"    a    ",
			"   sgs   ",
			"  msgsm  ",
			" mmmsmmm ",
			"sssssssss",
		},
	},
	"dyson_scaffold": {
		Art: []string{
			" ╭─┬─┬─╮ ",
			"╱╲ ╲ ╱ ╱╲",
			"│ ╲ ☼ ╱ │",
			"╲╱ ╱ ╲ ╲╱",
			" ╰─┴─┴─╯ ",
		},
		Tone: []string{
			" mmmmmmm ",
			"mmm m mmm",
			"m m f m m",
			"mmm m mmm",
			" mmmmmmm ",
		},
	},
}
