package templates

// StarterTemplate is one entry from the bundled starter template library —
// pure text art-direction prompts (composition, lighting, layout) with no
// reference image, transcribed from the user-provided template library.
// Category assignments are our own best-effort mapping onto each template's
// described use case, since the source material lists categories generally
// rather than tagging each template explicitly.
type StarterTemplate struct {
	Title    string
	Category string
	Prompt   string
}

// aspectRatioHint is appended to every starter template's prompt. The source
// material only gives an illustrative aspect-ratio example for one template,
// not a distinct value per template, so a single shared, commonly-used set is
// used rather than inventing per-template values that aren't actually there.
const aspectRatioHint = " Recommended aspect ratios: 4:5, 1:1, 2:3."

// StarterLibrary returns the 40 bundled starter templates.
func StarterLibrary() []StarterTemplate {
	return []StarterTemplate{
		{
			Title:    "Radial Festival Sunburst",
			Category: "Indian Festival",
			Prompt:   "Premium Indian festival poster composition built around a large glowing circular light source positioned behind the main visual. Strong radial symmetry with concentric rings and rays extending toward the edges. Central hero area with layered cultural decorations surrounding it. Rich warm gradients, atmospheric particles, ornamental Indian motifs, decorative borders and cinematic depth. High visual density without becoming cluttered. Strong central hierarchy for headline, hero visual and supporting information. Warm directional backlighting, glowing highlights and deep shadows. Premium commercial festival advertising aesthetic, photorealistic digital composite, sophisticated print-ready composition." + aspectRatioHint,
		},
		{
			Title:    "Royal Heritage",
			Category: "Cultural",
			Prompt:   "Luxury Indian heritage poster composition with a grand symmetrical architectural arch framing the central content. Palace-inspired pillars, carved ornamental details, paisley motifs, floral engraving and sophisticated decorative borders. Central focal area reserved for the primary visual and headline. Deep rich background with warm illumination and subtle atmospheric depth. Gold ornamental details, elegant symmetry and refined visual hierarchy. Dramatic warm spotlight from above-center with soft rim lighting. Premium royal editorial advertising aesthetic." + aspectRatioHint,
		},
		{
			Title:    "Black Gold Luxury",
			Category: "Luxury",
			Prompt:   "Premium luxury poster composition using a deep matte dark background with refined gold ornamental elements. Thin elegant frame around the composition, sophisticated decorative corners, subtle geometric patterns and controlled atmospheric particles. Large central hero area with dramatic spotlighting. Strong headline zone and clean supporting information areas. High contrast between deep shadows and luminous gold highlights. Minimal but expensive-looking composition. Luxury fashion, corporate and premium-brand advertising aesthetic." + aspectRatioHint,
		},
		{
			Title:    "Modern Corporate Grid",
			Category: "Corporate",
			Prompt:   "Clean modern corporate poster using a professional modular grid. Strong alignment, structured content zones, large hero visual area, clear headline hierarchy and carefully positioned supporting information. Minimal geometric decorations, subtle gradients, elegant divider lines and restrained visual effects. Balanced negative space. Contemporary sans-serif visual language, polished commercial photography, professional lighting and premium corporate branding aesthetic." + aspectRatioHint,
		},
		{
			Title:    "Diagonal Energy",
			Category: "Promotional",
			Prompt:   "Dynamic contemporary poster composition built around a strong diagonal visual direction from bottom-left toward top-right. Layered color washes, energetic brush strokes, angular geometric forms and overlapping translucent panels create movement. Main hero visual follows the diagonal axis. Bold headline area integrated into the movement of the composition. Dramatic lighting, particles, depth and controlled motion effects. High-impact modern advertising aesthetic." + aspectRatioHint,
		},
		{
			Title:    "Circular Mandala",
			Category: "Cultural",
			Prompt:   "Highly detailed symmetrical poster composition centered around a large circular mandala. Multiple concentric decorative rings, intricate Indian geometric patterns, floral geometry and radial linework surround the central hero area. Strong radial hierarchy with the most important visual content at the center. Decorative elements gradually become smaller toward the outer edges. Rich cultural detailing, balanced symmetry, refined ornamentation and premium print texture." + aspectRatioHint,
		},
		{
			Title:    "Grand Celebration",
			Category: "Event",
			Prompt:   "Spectacular celebration poster filled with layered decorative elements, flowers, ornamental objects, glowing particles and atmospheric lighting. Large central hero scene surrounded by secondary decorative elements. Strong headline zone above or overlapping the hero. Information zones arranged beneath the main visual. Rich depth between foreground, midground and background. Vibrant colors, dramatic illumination and abundant visual details. Premium large-scale celebration advertising aesthetic." + aspectRatioHint,
		},
		{
			Title:    "Editorial Magazine",
			Category: "Promotional",
			Prompt:   "High-end editorial magazine composition using an asymmetric professional grid. Large photographic hero visual with intentional overlap into typography. Oversized headline area, small supporting labels, elegant divider rules and carefully controlled whitespace. Sophisticated serif and sans-serif typographic contrast. Image and typography interact as one composition. Premium fashion magazine, lifestyle editorial and contemporary advertising aesthetic." + aspectRatioHint,
		},
		{
			Title:    "Ultra Minimal",
			Category: "Minimal",
			Prompt:   "Ultra-clean premium poster composition with a simple sophisticated background and one dominant hero visual. Large elegant headline with generous spacing. Very limited decorative elements. Carefully controlled whitespace, precise alignment, subtle geometric lines and soft shadows. Restrained color palette and refined visual hierarchy. Contemporary luxury branding aesthetic with professional studio lighting and clean commercial photography." + aspectRatioHint,
		},
		{
			Title:    "Product Hero",
			Category: "Product",
			Prompt:   "Premium commercial product advertising composition with the product as the dominant central visual. Dramatic studio lighting, controlled reflections, realistic shadows and subtle atmospheric effects. Background designed specifically to support the product without competing with it. Large headline zone above or beside the product. Supporting benefit information and call-to-action arranged beneath the hero. High-end commercial photography and luxury advertising aesthetic." + aspectRatioHint,
		},
		{
			Title:    "Futuristic Technology",
			Category: "Technology",
			Prompt:   "Futuristic technology poster with a dark atmospheric gradient background, glowing abstract 3D forms, subtle digital grids, luminous particles and translucent geometric panels. Central hero visual surrounded by volumetric light. Bold modern headline integrated with the futuristic environment. Layered depth using foreground objects, midground hero and distant digital structures. Neon accents, cinematic lighting and premium technology conference advertising aesthetic." + aspectRatioHint,
		},
		{
			Title:    "Traditional Kolam",
			Category: "Cultural",
			Prompt:   "Premium Indian cultural poster built around an intricate geometric kolam composition. Symmetrical chalk-like linework, radial petals, floral geometry and repeating decorative patterns form the visual foundation. Central hero area sits above or within the kolam structure. Traditional decorative objects may appear around the edges. Warm natural materials, subtle powder texture, parchment-like surfaces and soft golden lighting. Sophisticated Indian cultural advertising aesthetic." + aspectRatioHint,
		},
		{
			Title:    "Cinematic Hero",
			Category: "Cinematic",
			Prompt:   "Dramatic cinematic poster composition with one dominant hero visual occupying the central area. Strong backlight creates a luminous rim around the subject. Atmospheric haze, volumetric rays, floating particles and deep cinematic shadows create depth. Large headline integrated into the scene without obscuring the hero. Foreground, midground and background layers create a film-poster-like sense of scale. Premium cinematic advertising aesthetic." + aspectRatioHint,
		},
		{
			Title:    "Sunrise Inspiration",
			Category: "Seasonal",
			Prompt:   "Inspirational poster composition centered around a dramatic sunrise. Large glowing sun positioned behind the primary visual with long rays spreading through the composition. Warm atmospheric sky transitioning from deep tones into luminous gold. Hero subject placed against the light with subtle rim illumination. Strong headline area positioned around the sunrise focal point. Cinematic haze, clouds, particles and layered depth. Optimistic premium campaign aesthetic." + aspectRatioHint,
		},
		{
			Title:    "Floral Frame",
			Category: "Wedding",
			Prompt:   "Elegant poster composition surrounded by a rich botanical frame. Flowers, leaves and decorative vines naturally enter from all four corners while leaving a strong central visual area. Central hero visual and headline receive the highest contrast. Fine ornamental borders, subtle botanical patterns and soft atmospheric lighting. Sophisticated symmetry with organic variation. Premium wedding, celebration, lifestyle and cultural advertising aesthetic." + aspectRatioHint,
		},
		{
			Title:    "Neon Night",
			Category: "Event",
			Prompt:   "Modern nightlife poster composition with a deep dark background, vivid neon illumination and atmospheric haze. Central hero area surrounded by glowing geometric forms, reflective surfaces, light trails and floating particles. Large bold headline integrated into the luminous environment. Strong contrast, vibrant highlights and cinematic depth. Contemporary entertainment, nightlife and youth-event advertising aesthetic." + aspectRatioHint,
		},
		{
			Title:    "Diya Light Festival",
			Category: "Indian Festival",
			Prompt:   "Indian light-festival poster composition based around a powerful central warm light source. Multiple glowing lamps distributed through the foreground and side areas create depth. Decorative rangoli-inspired patterns, floral ornaments and traditional motifs surround the main visual. Darker background transitions toward brilliant golden illumination around the hero. Strong warm highlights, soft haze and glowing particles. Premium Indian celebration advertising aesthetic." + aspectRatioHint,
		},
		{
			Title:    "Organic Nature",
			Category: "Seasonal",
			Prompt:   "Nature-inspired poster composition using flowing organic shapes and botanical elements. Leaves and branches enter naturally from the edges and frame the central visual. Soft natural sunlight, atmospheric depth, subtle mist and realistic textures. Earthy tones and elegant curved divisions create a calm visual rhythm. Central headline and hero visual remain highly readable. Premium environmental, wellness and lifestyle campaign aesthetic." + aspectRatioHint,
		},
		{
			Title:    "Newspaper Editorial",
			Category: "Promotional",
			Prompt:   "Professional news and editorial poster using a structured rectangular grid. Strong headline block, large central image panel, secondary information panels and compact supporting elements. Sharp divider lines, bold typographic hierarchy and controlled spacing. Subtle paper texture and restrained color palette. Authoritative editorial visual language inspired by premium newspapers and modern news publications." + aspectRatioHint,
		},
		{
			Title:    "Luxury Gold Frame",
			Category: "Luxury",
			Prompt:   "Elegant premium poster surrounded by a large ornamental frame with sophisticated gold detailing. Decorative corners, subtle embossed patterns and refined geometric ornamentation frame the central composition. Central hero area receives dramatic warm illumination. Headline and supporting information follow a symmetrical hierarchy. Dark or neutral background creates contrast against the decorative frame. Classical luxury invitation and premium celebration advertising aesthetic." + aspectRatioHint,
		},
		{
			Title:    "Soft Pastel",
			Category: "Community",
			Prompt:   "Elegant contemporary poster using a soft pastel color environment with subtle gradients and diffused lighting. Rounded organic shapes, delicate decorative elements and gentle shadows create a calm premium atmosphere. Central hero image surrounded by generous breathing room. Refined typography hierarchy with soft visual contrast. Suitable for lifestyle, beauty, education, wellness, family and modern event campaigns." + aspectRatioHint,
		},
		{
			Title:    "Bold Typography",
			Category: "Promotional",
			Prompt:   "Typography-first poster composition where the headline is the dominant visual element. Oversized expressive typography occupies most of the central composition. Hero imagery is integrated behind, inside or around the letterforms. Strong geometric blocks, contrasting shapes and dynamic alignment create visual energy. Supporting information is kept compact. Modern art-direction, high-impact advertising and contemporary campaign aesthetic." + aspectRatioHint,
		},
		{
			Title:    "Split Screen",
			Category: "Business",
			Prompt:   "Modern poster divided into two visually connected sections. One section contains the main hero image while the other contains the primary headline and supporting information. Strong vertical or horizontal division with subtle overlap between both sides. Consistent color system connects the sections. Clean typography, geometric separators, balanced proportions and premium commercial photography. Contemporary corporate, event and promotional advertising aesthetic." + aspectRatioHint,
		},
		{
			Title:    "Circular Portrait",
			Category: "Professional Services",
			Prompt:   "Premium portrait poster composition centered around a large circular portrait or hero image. Circular image frame surrounded by subtle rings, geometric ornaments and atmospheric glow. Headline positioned around or beneath the portrait. Supporting information follows a balanced vertical hierarchy. Elegant background gradients, controlled particles and refined decorative details. Suitable for professional profiles, leadership campaigns, celebrations and promotional posters." + aspectRatioHint,
		},
		{
			Title:    "Luxury Marble",
			Category: "Luxury",
			Prompt:   "Luxury poster composition using a sophisticated marble-inspired background with subtle natural veins and premium texture. Central hero visual positioned against the marble surface with dramatic directional lighting. Thin geometric frame, refined typography, restrained ornamental details and subtle metallic-looking accents. High-end architectural and luxury-brand visual language. Elegant, clean and expensive-looking composition." + aspectRatioHint,
		},
		{
			Title:    "Paper Craft",
			Category: "Promotional",
			Prompt:   "Creative layered paper-craft poster composition where background, decorative elements and hero objects appear physically constructed from overlapping paper layers. Visible paper edges, subtle shadows, cutout shapes and tactile material texture. Central headline integrated into the layered paper environment. Rich depth created through foreground and background paper elements. Warm artistic lighting and sophisticated handcrafted advertising aesthetic." + aspectRatioHint,
		},
		{
			Title:    "3D Isometric",
			Category: "Technology",
			Prompt:   "Modern 3D isometric poster composition featuring a miniature stylized environment viewed from an elevated isometric perspective. Central hero scene sits on a designed platform with surrounding objects arranged around it. Floating typography and geometric information panels create hierarchy. Soft studio lighting, realistic materials, ambient occlusion and subtle shadows. Premium contemporary 3D advertising aesthetic." + aspectRatioHint,
		},
		{
			Title:    "Collage",
			Category: "Cultural",
			Prompt:   "Editorial artistic collage poster combining multiple photographic elements, textures, typography fragments, geometric shapes and layered paper-like surfaces. Intentional overlaps create depth and visual storytelling. One dominant focal area anchors the composition while secondary visuals surround it. Controlled asymmetry, sophisticated color relationships and premium art-direction. Contemporary cultural, fashion and creative campaign aesthetic." + aspectRatioHint,
		},
		{
			Title:    "Geometric Modern",
			Category: "Corporate",
			Prompt:   "Modern geometric poster composition built from circles, rectangles, diagonal planes and clean linear structures. Strong modular grid with deliberate overlaps between geometric elements and the hero visual. Large headline positioned within or across geometric blocks. Controlled color palette, crisp edges, subtle shadows and sophisticated spacing. Contemporary corporate, technology and event advertising aesthetic." + aspectRatioHint,
		},
		{
			Title:    "Traditional Textile",
			Category: "Cultural",
			Prompt:   "Indian-inspired poster composition using traditional textile patterns as the visual foundation. Rich woven textures, repeating motifs, borders, paisley forms and ornamental patterns frame the central content. Central hero visual receives strong contrast against the textile background. Layered decorative depth, warm directional lighting and tactile fabric texture. Premium cultural, wedding, festival and heritage advertising aesthetic." + aspectRatioHint,
		},
		{
			Title:    "Monochrome Professional",
			Category: "Professional Services",
			Prompt:   "Sophisticated monochromatic poster composition using one dominant color family with controlled tonal variations. Strong hero image, bold headline and structured supporting information. Minimal geometric accents, elegant dividers and subtle gradients. High contrast between light and dark tones creates visual depth without relying on multiple colors. Premium professional, corporate and executive campaign aesthetic." + aspectRatioHint,
		},
		{
			Title:    "Dynamic Motion Blur",
			Category: "Fitness",
			Prompt:   "High-energy poster composition designed to visually communicate speed and movement. Hero subject appears sharp while the surrounding environment contains controlled directional motion blur, light trails and energetic streaks. Strong diagonal or curved composition. Bold headline integrated into the movement. Dramatic lighting, high contrast and cinematic depth. Suitable for sports, automotive, youth, entertainment and action campaigns." + aspectRatioHint,
		},
		{
			Title:    "Festive Rangoli",
			Category: "Indian Festival",
			Prompt:   "Indian festive poster centered around an elaborate colorful rangoli pattern. Radial floral geometry forms a strong visual anchor beneath or behind the main hero. Decorative lamps, flowers and cultural objects can frame the composition. Powder texture, chalk edges and realistic surface imperfections create tactile authenticity. Warm festive lighting, rich cultural colors and layered depth. Premium Indian festival advertising aesthetic." + aspectRatioHint,
		},
		{
			Title:    "Temple Architecture",
			Category: "Cultural",
			Prompt:   "Grand Indian temple-inspired poster composition with architectural pillars, ornamental arches, sculptural details and symmetrical framing. Central visual area positioned like a sacred focal point beneath the architectural structure. Intricate carvings, warm stone textures, lamps and atmospheric golden light create depth. Strong vertical hierarchy and refined decorative borders. Premium heritage and cultural campaign aesthetic." + aspectRatioHint,
		},
		{
			Title:    "Celebration Confetti",
			Category: "Event",
			Prompt:   "Joyful modern celebration poster with a bright central hero area surrounded by floating confetti, ribbons, particles and decorative shapes. Large headline dominates the upper-middle area while supporting information occupies the lower section. Layered foreground and background elements create depth. Vibrant but controlled colors, bright studio lighting and polished commercial finish. Premium birthday, achievement, anniversary and celebration advertising aesthetic." + aspectRatioHint,
		},
		{
			Title:    "Professional Services",
			Category: "Professional Services",
			Prompt:   "Trust-focused professional services poster using a clean structured composition. Large professional hero image or subject on one side and strong headline/information hierarchy on the other. Subtle geometric background, restrained colors, clean divider lines and sophisticated typography. Strong emphasis on clarity, credibility and readability. Premium financial, consulting, employment, legal, education and business-services advertising aesthetic." + aspectRatioHint,
		},
		{
			Title:    "Real Estate Luxury",
			Category: "Real Estate",
			Prompt:   "Premium real-estate poster composition featuring a large architectural hero image with cinematic perspective. Elegant dark or neutral background, sophisticated typography, thin geometric frame and refined information blocks. Warm architectural lighting, realistic materials and subtle atmospheric depth. Property image remains dominant while text stays highly readable. Luxury real-estate marketing aesthetic." + aspectRatioHint,
		},
		{
			Title:    "Food Photography",
			Category: "Food",
			Prompt:   "Premium food advertising poster centered around an appetizing hero food composition. Hero dish photographed from a dramatic close or three-quarter angle with realistic texture, steam and controlled studio lighting. Supporting ingredients arranged naturally around the hero. Warm atmospheric background, subtle decorative elements and strong headline hierarchy. High-end restaurant and food-brand advertising aesthetic." + aspectRatioHint,
		},
		{
			Title:    "Fitness Energy",
			Category: "Fitness",
			Prompt:   "High-energy fitness poster composition with a powerful athletic hero subject positioned dynamically in the center. Dramatic directional lighting, strong shadows, subtle sweat and realistic muscle definition. Energetic geometric shapes, particles and directional streaks create movement. Bold typography and compact supporting information. Strong contrast and motivational advertising aesthetic." + aspectRatioHint,
		},
		{
			Title:    "Education / Academic",
			Category: "Education",
			Prompt:   "Professional education poster using a clean academic composition. Central student, educator or educational environment as the hero visual. Structured headline area, supporting information blocks, subtle geometric academic motifs and organized spacing. Bright professional lighting, approachable visual tone and high readability. Premium school, college, training, workshop and educational campaign aesthetic." + aspectRatioHint,
		},
	}
}
