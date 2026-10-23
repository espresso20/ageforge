package game

// The save the storage wall has to load: a Bronze Age town built the way the
// old rules allowed and a patient player built it. It holds every storage
// copy the old limits sold (50 Stashes, 25 Storage Pits, 25 Warehouses), 20
// Forager Posts, stores full to the old caps in wood and stone and a large
// knowledge stock, and a plan whose next Warehouse and next House had already
// banked part of their price from overflow.
//
// The save was written by the game as it stood before the wall (a checkout
// from then regenerates it); the committed bytes are what the old game wrote
// and signed.
const storageWallFixture = "testdata/pre_storage_wall_bronze.json"
