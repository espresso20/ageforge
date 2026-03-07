Layout strategy specs (implement each):

       placeBuildingsOrganic (primitive/stone):
       - Random scatter within an ellipse radius ~40% of map size
       - Plot cell size 12px, minimum 3px padding
       - No roads or structure — pure organic scatter
       - Groups of 3-5 huts near each other with 8-20px between groups
       - Use random walk from center, each building placed within 15px of a random existing building

       placeBuildingsVillage (bronze/iron/classical):
       - Place the most-assigned building (or a shrine/temple type) at center
       - Radiate outward in 4-6 "spokes" rather than concentric rings
       - Each spoke is a road: draw a 1-2px wide line of road pixels outward, buildings placed on either side of it
       - Road pixel color: slightly lighter than terrain, desaturated
       - Max radius ~35% of map, plot cell 10px

       placeBuildingsMedieval (medieval/renaissance):
       - Place the highest-tier military building at center (or center of map) as a "castle"
       - Draw 2px thick walls as a rough square/circle around it (~30px radius)
       - Divide remaining buildings into 4 quarters (NW/NE/SW/SE)
       - Each quarter fills with its buildings in a loose grid, 10px spacing
       - Draw 1px roads connecting quarters through the center

       placeBuildingsIndustrial (colonial/industrial):
       - Divide city into zones: production zone (left half), residential zone (right half)
       - Production buildings (food/lumber/masonry/metallurgy) go left
       - Population/knowledge/trade buildings go right
       - Strict grid: 14px cell spacing, perfectly aligned rows and columns
       - Draw road grid: horizontal road every 3 cells, vertical every 4 cells (1px, slightly lighter terrain color)

       placeBuildingsModern (modern/atomic):
       - City blocks: 6×4 grids of buildings separated by 3px avenues
       - Start placing from top-left of city area, fill block by block
       - Between blocks, draw 2px streets (lighter paved color)
       - Wonders placed in the center "civic district" block

       placeBuildingsCampus (digital/nano):
       - Distinct campus clusters of 6-10 buildings each
       - Each cluster is a tight circle with 8px cell size
       - 20-30px gap between campus clusters
       - Winding path (1px) connects clusters
       - Cluster centers arranged in a loose hexagonal pattern

       placeBuildingsOrbital (space/galactic/cosmic):
       - Central hub structure at dead center (largest wonder or most-assigned building)
       - 3 orbital rings at radii 25px, 45px, 70px from center
       - Buildings placed along rings with equal angular spacing
       - Draw the rings as dotted circles (draw every 3rd pixel on ring)
       - Colors: blue/purple tones for paths

       For all layouts: if a position is occupied (grid.isFree returns false), try up to 20 random offsets before skipping the building. Do not place buildings within 8px of image
        edges or within water/deep-water tiles (check terrain color).

       2. Building Sprite System

       Replace generic shape rendering with category-aware pixel sprites. Buildings should be 6-12px wide × 6-12px tall (full map) or 3-6px (minimap).

       Add a buildingCategory function that determines the visual type:

       type spriteType int

       const (
           spriteHut spriteType = iota
           spriteFarm
           spriteMill
           spriteLumberCamp
           spriteMine
           spriteFortress
           spriteBarracks
           spriteTemple
           spriteLibrary
           spriteMarket
           spriteFactory
           spriteWorkshop
           spritePalace
           spriteObservatory
           spriteDome
           spriteSkyscraper
           spriteServer
           spriteSpaceStation
           spriteWonder
       )

       Add func getBuildingSprite(domain string, buildingKey string, era string) spriteType that maps domain+era to sprite type:
       - domain "food" + primitive/stone → spriteHut
       - domain "food" + later → spriteFarm
       - domain "lumber" → spriteLumberCamp
       - domain "masonry" → spriteMine
       - domain "military" + early → spriteBarracks, + late → spriteFortress
       - domain "knowledge" + early → spriteLibrary, + late → spriteObservatory, + space → spriteServer
       - domain "faith" → spriteTemple
       - domain "trade" → spriteMarket
       - domain "engineering" or "metallurgy" → spriteFactory (industrial), spriteWorkshop (early)
       - domain "energy" + early → spriteWorkshop, + late → spriteDome
       - domain "hacker" → spriteServer
       - domain "astronaut" → spriteSpaceStation
       - isWonder → spriteWonder (always)

       Then implement a drawBuildingSprite(img *image.RGBA, px, py int, stype spriteType, primary, accent color.RGBA, scale int) function. Scale 1 = minimap (3-5px), scale 2 =
       full map (6-12px).

       Pixel art sprites (scale 2, so multiply each pixel coordinate by scale):

       spriteHut (5×5 base):
         .P.
        .PPP.
        PPPPP
        A.A.A
       P = primary color, A = accent color (lighter shade), . = transparent (don't draw)

       spriteFarm (7×5 base):
       AAAAAA A  (top: field rows, alternating primary and accent horizontal lines)
       PPPPPP P
       AAAAAA A
       PPPPPP P
       ..III..   (bottom: fence posts as thin vertical lines in accent)
       Draw horizontal stripes alternating primary/accent, then draw fence-like bottom

       spriteLumberCamp (5×7 base):
         .A.
         .A.
        PPPPP
        P.P.P
        PPPPP
        .PPP.
        .....
       Tall trunk (A) above a structure (P)

       spriteMine (6×5 base):
        PPPPPP
        P.AAAP  (arch entrance: side walls P, inside accent curve)
        P.....P  (open entrance)
        AAAAAAA  (ground line)
       Draw arch shape — semicircle cutout in rectangular block

       spriteBarracks (7×5 base):
       PPPPPPP  (flat roof)
       P.P.P.P  (windows: alternating primary/transparent)
       PPPPPPP  (main wall)
       AAAAAAA  (foundation, slightly lighter)

       spriteFortress (7×7 base):
       P.P P.P  (battlements: square teeth on top)
       PPPPPPP  (parapet)
       P.....P  (arrow slit)
       PPPPPPP  (main wall)
       A.AAA.A  (gate arch — accent color on sides, gap in middle)

       spriteTemple (7×8 base):
         .A.    (spire tip)
         AAA    (upper spire)
        AAAAA   (spire base)
        PPPPP   (column tops)
        P.P.P   (columns with gaps)
        PPPPP   (floor)
       AAAAAAA  (steps)

       spriteLibrary (7×6 base):
       AAAAAAA  (roof)
       PPPPPPP  (upper wall)
       P.P.P.P  (windows)
       PPPPPPP  (mid wall)
       P.....P  (entry)
       AAAAAAA  (step)

       spriteMarket (7×5 base):
       .AAAAA.  (awning/overhang)
       PPPPPPP  (front wall with awning supports)
       P.P.P.P  (market stalls: alternating)
       PPPPPPP  (base)
       AAAAAAA  (ground)

       spriteFactory (8×7 base):
       .A..A..  (chimney tops)
       AA..AA.  (chimneys)
       AA..AA.  (chimneys)
       PPPPPPPP (main building roof)
       PPPPPPPP (main wall)
       P.PP.PP. (windows/doors)
       AAAAAAAA (foundation)

       spriteWorkshop (6×5 base):
       .PPPP.   (slanted roof left)
       PPPPPP   (roof base)
       P.PP.P   (walls with window)
       PPPPPP   (lower wall)
       AAAAAA   (foundation)

       spriteObservatory (7×7 base):
         .A.    (telescope tip)
         AAA    (dome top)
        AAAAA   (dome mid)
        AAAAA   (dome lower)
        PPPPP   (base ring)
        P...P   (column)
       PPPPPPP  (foundation)

       spriteDome (8×6 base):
        .AAAA.  (dome top)
       .AAAAAA. (dome wide)
       AAAAAAAA (dome lower)
       PPPPPPPP (base)
       PP....PP (airlock/entry)
       PPPPPPPP (ground)

       spriteServer (5×7 base):
       PPPPP    (rack top)
       APPPA    (server unit)
       PPPPP    (gap)
       APPPA    (server unit)
       PPPPP    (gap)
       APPPA    (server unit)
       AAAAA    (base)

       spriteSkyscraper (5×10 base):
         A      (antenna)
        PAP     (upper floor)
        PPP     (upper floors ×3)
       PPPPP    (mid section)
       P.P.P    (windows ×2)
       PPPPP    (lower)
       AAAAA    (lobby/base)

       spriteSpaceStation (9×7 base):
       ....A....  (central hub top)
       ...AAA...  (hub)
       AA.AAA.AA  (solar panels: A on far sides, hub center)
       AA.AAA.AA  (panels)
       ...AAA...  (hub)
       ....A....  (hub bottom)
       .........  (space below)

       spriteWonder (10×10 base, always largest on map):
       Design a generic "impressive structure" — use both colors, add height, make it clearly bigger and more detailed than normal buildings. Draw a stepped pyramid shape or
       palace shape.

       For the minimap (scale=1), just draw a 3×3 or 4×4 colored block using primary color — detailed sprites are too small to matter there.

       Implement the sprite drawing by iterating pixel rows/columns and calling setPixel(img, px+col*scale, py+row*scale, color) for each pixel in the pattern. Use primary for P
       pixels, accent for A pixels, and skip . pixels.

       3. Wonder Scaling

       Wonders should be rendered at 2× the normal size (scale 3 on full map). Mark buildings as wonders if their domain key equals "" or their key appears in config.Wonders().

       4. Keep Road Drawing

       Keep any existing road/path drawing code. If none exists, add light terrain-colored paths (slightly lighter than surrounding terrain) connecting buildings in the
       village/medieval/industrial layouts as described above.

       Implementation Notes

       - Check the actual struct/function names in mapgen.go — adapt the above spec to fit the existing code structure rather than doing a complete from-scratch rewrite
       - If buildingPlacement has different field names, use the actual names
       - If the current code passes buildings as a different type, adapt accordingly
       - The PlotGrid collision system is new — integrate it into each layout function
       - Make sure to import "math" for trig functions needed by orbital layout (sin, cos)
       - All layout functions should call drawBuildingSprite instead of the old shape-based drawing