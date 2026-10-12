package config

// The town rule: how many copies of a building the town an age is priced
// for holds.
//
// Two numbers, not one. StorageRuleCopies is the copies of an age's storage
// building the reference town holds (the storage rule, and the storage
// buildings a gate asks for). TownCopies is the copies of every other
// building: the gate rule's count, and the count the storage rule prices its
// dearest ask at. Storage climbs at StorageRate and is a wall; everything
// else climbs at ProducerRate and is bought for what it makes, which is why
// the two are sized apart.
//
// Where TownCopies comes from. A building that makes the same amount in every
// copy and whose price climbs by r a copy costs P r^(n-1) for copy n, and the
// n-1 copies owned already make (n-1) o between them. The wait to afford copy
// n is therefore P r^(n-1) / ((n-1) o): proportional to r^(n-1) / (n-1), with
// r the rate producers climb at (ProducerRate). Copy n+1 is a shorter wait
// than copy n exactly when r (n-1) < n, that is for n-1 below 1/(r-1) = 6.67,
// and a longer one after. So the wait is shortest at the copy with seven
// before it, the eighth: past it each further copy costs more of the town's
// time than the one before. TestTownCopiesIsTheShortestWait works it out from
// ProducerRate and asserts it.
const (
	// ProducerRate is the rate a building that makes something climbs at for
	// each copy owned: the factor on every producer, research and military
	// building in the tables (a test holds them to it).
	ProducerRate = 1.15
	// TownCopies is the copies of every building that is not storage in the
	// town an age is priced for.
	TownCopies = 8
)
