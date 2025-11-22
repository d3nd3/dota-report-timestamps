const DOTA_HEROES = [
    { id: 'antimage', name: 'Anti-Mage', localized_name: 'Anti-Mage' },
    { id: 'axe', name: 'Axe', localized_name: 'Axe' },
    { id: 'bane', name: 'Bane', localized_name: 'Bane' },
    { id: 'bloodseeker', name: 'Bloodseeker', localized_name: 'Bloodseeker' },
    { id: 'crystal_maiden', name: 'Crystal Maiden', localized_name: 'Crystal Maiden' },
    { id: 'drow_ranger', name: 'Drow Ranger', localized_name: 'Drow Ranger' },
    { id: 'earthshaker', name: 'Earthshaker', localized_name: 'Earthshaker' },
    { id: 'juggernaut', name: 'Juggernaut', localized_name: 'Juggernaut' },
    { id: 'mirana', name: 'Mirana', localized_name: 'Mirana' },
    { id: 'morphling', name: 'Morphling', localized_name: 'Morphling' },
    { id: 'nevermore', name: 'Shadow Fiend', localized_name: 'Shadow Fiend' },
    { id: 'phantom_lancer', name: 'Phantom Lancer', localized_name: 'Phantom Lancer' },
    { id: 'puck', name: 'Puck', localized_name: 'Puck' },
    { id: 'pudge', name: 'Pudge', localized_name: 'Pudge' },
    { id: 'razor', name: 'Razor', localized_name: 'Razor' },
    { id: 'sand_king', name: 'Sand King', localized_name: 'Sand King' },
    { id: 'storm_spirit', name: 'Storm Spirit', localized_name: 'Storm Spirit' },
    { id: 'sven', name: 'Sven', localized_name: 'Sven' },
    { id: 'tiny', name: 'Tiny', localized_name: 'Tiny' },
    { id: 'vengefulspirit', name: 'Vengeful Spirit', localized_name: 'Vengeful Spirit' },
    { id: 'windrunner', name: 'Windrunner', localized_name: 'Windrunner' },
    { id: 'zeus', name: 'Zeus', localized_name: 'Zeus' },
    { id: 'kunkka', name: 'Kunkka', localized_name: 'Kunkka' },
    { id: 'kez', name: 'Kez', localized_name: 'Kez' },
    { id: 'lina', name: 'Lina', localized_name: 'Lina' },
    { id: 'lion', name: 'Lion', localized_name: 'Lion' },
    { id: 'shadow_shaman', name: 'Shadow Shaman', localized_name: 'Shadow Shaman' },
    { id: 'slardar', name: 'Slardar', localized_name: 'Slardar' },
    { id: 'tidehunter', name: 'Tidehunter', localized_name: 'Tidehunter' },
    { id: 'witch_doctor', name: 'Witch Doctor', localized_name: 'Witch Doctor' },
    { id: 'lich', name: 'Lich', localized_name: 'Lich' },
    { id: 'riki', name: 'Riki', localized_name: 'Riki' },
    { id: 'enigma', name: 'Enigma', localized_name: 'Enigma' },
    { id: 'tinker', name: 'Tinker', localized_name: 'Tinker' },
    { id: 'sniper', name: 'Sniper', localized_name: 'Sniper' },
    { id: 'necrolyte', name: 'Necrophos', localized_name: 'Necrophos' },
    { id: 'warlock', name: 'Warlock', localized_name: 'Warlock' },
    { id: 'beastmaster', name: 'Beastmaster', localized_name: 'Beastmaster' },
    { id: 'queenofpain', name: 'Queen of Pain', localized_name: 'Queen of Pain' },
    { id: 'venomancer', name: 'Venomancer', localized_name: 'Venomancer' },
    { id: 'faceless_void', name: 'Faceless Void', localized_name: 'Faceless Void' },
    { id: 'skeleton_king', name: 'Wraith King', localized_name: 'Wraith King' },
    { id: 'death_prophet', name: 'Death Prophet', localized_name: 'Death Prophet' },
    { id: 'phantom_assassin', name: 'Phantom Assassin', localized_name: 'Phantom Assassin' },
    { id: 'pugna', name: 'Pugna', localized_name: 'Pugna' },
    { id: 'templar_assassin', name: 'Templar Assassin', localized_name: 'Templar Assassin' },
    { id: 'viper', name: 'Viper', localized_name: 'Viper' },
    { id: 'luna', name: 'Luna', localized_name: 'Luna' },
    { id: 'dragon_knight', name: 'Dragon Knight', localized_name: 'Dragon Knight' },
    { id: 'dazzle', name: 'Dazzle', localized_name: 'Dazzle' },
    { id: 'rattletrap', name: 'Clockwerk', localized_name: 'Clockwerk' },
    { id: 'leshrac', name: 'Leshrac', localized_name: 'Leshrac' },
    { id: 'furion', name: 'Nature\'s Prophet', localized_name: 'Nature\'s Prophet' },
    { id: 'life_stealer', name: 'Lifestealer', localized_name: 'Lifestealer' },
    { id: 'dark_seer', name: 'Dark Seer', localized_name: 'Dark Seer' },
    { id: 'clinkz', name: 'Clinkz', localized_name: 'Clinkz' },
    { id: 'omniknight', name: 'Omniknight', localized_name: 'Omniknight' },
    { id: 'enchantress', name: 'Enchantress', localized_name: 'Enchantress' },
    { id: 'huskar', name: 'Huskar', localized_name: 'Huskar' },
    { id: 'night_stalker', name: 'Night Stalker', localized_name: 'Night Stalker' },
    { id: 'broodmother', name: 'Broodmother', localized_name: 'Broodmother' },
    { id: 'bounty_hunter', name: 'Bounty Hunter', localized_name: 'Bounty Hunter' },
    { id: 'weaver', name: 'Weaver', localized_name: 'Weaver' },
    { id: 'jakiro', name: 'Jakiro', localized_name: 'Jakiro' },
    { id: 'batrider', name: 'Batrider', localized_name: 'Batrider' },
    { id: 'chen', name: 'Chen', localized_name: 'Chen' },
    { id: 'spectre', name: 'Spectre', localized_name: 'Spectre' },
    { id: 'ancient_apparition', name: 'Ancient Apparition', localized_name: 'Ancient Apparition' },
    { id: 'doom_bringer', name: 'Doom', localized_name: 'Doom' },
    { id: 'ursa', name: 'Ursa', localized_name: 'Ursa' },
    { id: 'spirit_breaker', name: 'Spirit Breaker', localized_name: 'Spirit Breaker' },
    { id: 'gyrocopter', name: 'Gyrocopter', localized_name: 'Gyrocopter' },
    { id: 'alchemist', name: 'Alchemist', localized_name: 'Alchemist' },
    { id: 'invoker', name: 'Invoker', localized_name: 'Invoker' },
    { id: 'silencer', name: 'Silencer', localized_name: 'Silencer' },
    { id: 'obsidian_destroyer', name: 'Outworld Devourer', localized_name: 'Outworld Devourer' },
    { id: 'lycan', name: 'Lycan', localized_name: 'Lycan' },
    { id: 'brewmaster', name: 'Brewmaster', localized_name: 'Brewmaster' },
    { id: 'shadow_demon', name: 'Shadow Demon', localized_name: 'Shadow Demon' },
    { id: 'lone_druid', name: 'Lone Druid', localized_name: 'Lone Druid' },
    { id: 'chaos_knight', name: 'Chaos Knight', localized_name: 'Chaos Knight' },
    { id: 'meepo', name: 'Meepo', localized_name: 'Meepo' },
    { id: 'treant', name: 'Treant Protector', localized_name: 'Treant Protector' },
    { id: 'ogre_magi', name: 'Ogre Magi', localized_name: 'Ogre Magi' },
    { id: 'undying', name: 'Undying', localized_name: 'Undying' },
    { id: 'rubick', name: 'Rubick', localized_name: 'Rubick' },
    { id: 'disruptor', name: 'Disruptor', localized_name: 'Disruptor' },
    { id: 'nyx_assassin', name: 'Nyx Assassin', localized_name: 'Nyx Assassin' },
    { id: 'naga_siren', name: 'Naga Siren', localized_name: 'Naga Siren' },
    { id: 'keeper_of_the_light', name: 'Keeper of the Light', localized_name: 'Keeper of the Light' },
    { id: 'wisp', name: 'Io', localized_name: 'Io' },
    { id: 'visage', name: 'Visage', localized_name: 'Visage' },
    { id: 'slark', name: 'Slark', localized_name: 'Slark' },
    { id: 'medusa', name: 'Medusa', localized_name: 'Medusa' },
    { id: 'troll_warlord', name: 'Troll Warlord', localized_name: 'Troll Warlord' },
    { id: 'centaur', name: 'Centaur Warrunner', localized_name: 'Centaur Warrunner' },
    { id: 'magnataur', name: 'Magnus', localized_name: 'Magnus' },
    { id: 'shredder', name: 'Timbersaw', localized_name: 'Timbersaw' },
    { id: 'bristleback', name: 'Bristleback', localized_name: 'Bristleback' },
    { id: 'tusk', name: 'Tusk', localized_name: 'Tusk' },
    { id: 'skywrath_mage', name: 'Skywrath Mage', localized_name: 'Skywrath Mage' },
    { id: 'abaddon', name: 'Abaddon', localized_name: 'Abaddon' },
    { id: 'elder_titan', name: 'Elder Titan', localized_name: 'Elder Titan' },
    { id: 'legion_commander', name: 'Legion Commander', localized_name: 'Legion Commander' },
    { id: 'techies', name: 'Techies', localized_name: 'Techies' },
    { id: 'ember_spirit', name: 'Ember Spirit', localized_name: 'Ember Spirit' },
    { id: 'earth_spirit', name: 'Earth Spirit', localized_name: 'Earth Spirit' },
    { id: 'abyssal_underlord', name: 'Underlord', localized_name: 'Underlord' },
    { id: 'terrorblade', name: 'Terrorblade', localized_name: 'Terrorblade' },
    { id: 'phoenix', name: 'Phoenix', localized_name: 'Phoenix' },
    { id: 'oracle', name: 'Oracle', localized_name: 'Oracle' },
    { id: 'winter_wyvern', name: 'Winter Wyvern', localized_name: 'Winter Wyvern' },
    { id: 'arc_warden', name: 'Arc Warden', localized_name: 'Arc Warden' },
    { id: 'monkey_king', name: 'Monkey King', localized_name: 'Monkey King' },
    { id: 'dark_willow', name: 'Dark Willow', localized_name: 'Dark Willow' },
    { id: 'pangolier', name: 'Pangolier', localized_name: 'Pangolier' },
    { id: 'grimstroke', name: 'Grimstroke', localized_name: 'Grimstroke' },
    { id: 'hoodwink', name: 'Hoodwink', localized_name: 'Hoodwink' },
    { id: 'void_spirit', name: 'Void Spirit', localized_name: 'Void Spirit' },
    { id: 'snapfire', name: 'Snapfire', localized_name: 'Snapfire' },
    { id: 'mars', name: 'Mars', localized_name: 'Mars' },
    { id: 'dawnbreaker', name: 'Dawnbreaker', localized_name: 'Dawnbreaker' },
    { id: 'marci', name: 'Marci', localized_name: 'Marci' },
    { id: 'primal_beast', name: 'Primal Beast', localized_name: 'Primal Beast' },
    { id: 'muerta', name: 'Muerta', localized_name: 'Muerta' },
    { id: 'ringmaster', name: 'Ringmaster', localized_name: 'Ringmaster' }
];

    function convertSteamIDTo64(steamID) {
        if (!steamID) return null;
        const steamIDStr = String(steamID).trim();
        const STEAMID64_IDENTIFIER = BigInt('76561197960265728');
        
        if (steamIDStr.startsWith('[') && steamIDStr.endsWith(']')) {
            const match = steamIDStr.match(/\[U:1:(\d+)\]/);
            if (match) {
                const accountID = BigInt(match[1]);
                const steamID64 = accountID + STEAMID64_IDENTIFIER;
                return steamID64.toString();
            }
        } else if (steamIDStr.startsWith('U:1:')) {
            const accountID = BigInt(steamIDStr.substring(4));
            const steamID64 = accountID + STEAMID64_IDENTIFIER;
            return steamID64.toString();
        } else if (/^\d+$/.test(steamIDStr)) {
            const numID = BigInt(steamIDStr);
            if (numID < STEAMID64_IDENTIFIER) {
                const steamID64 = numID + STEAMID64_IDENTIFIER;
                return steamID64.toString();
            }
        }
        
        return steamIDStr;
    }

    function getSlotColor(slot) {
        if (slot === null || slot === undefined || slot < 0 || slot >= 10) {
            return 'var(--text-primary)';
        }
        const slotColors = [
            '#3375FF', // Slot 0 - Blue
            '#66FFBF', // Slot 1 - Aquamarine
            '#BF00BF', // Slot 2 - Purple
            '#F3F00B', // Slot 3 - Yellow
            '#FF6B00', // Slot 4 - Orange
            '#FE86C2', // Slot 5 - Pink
            '#A1B447', // Slot 6 - Olive
            '#65D9F7', // Slot 7 - Sky Blue
            '#008321', // Slot 8 - Green
            '#A46900'  // Slot 9 - Brown
        ];
        return slotColors[slot];
    }

    function getHeroIconUrl(heroName) {
    if (!heroName) return null;
    
    const normalizeHeroName = (name) => {
        return name
            .replace(/([a-z])([A-Z])/g, '$1_$2')
            .toLowerCase()
            .replace(/\s+/g, '_');
    };
    
    const heroNameVariations = {
        'zuus': 'zeus',
        'zeus': 'zuus'
    };
    
    let heroNameNormalized = normalizeHeroName(heroName);
    if (heroNameVariations[heroNameNormalized]) {
        heroNameNormalized = heroNameVariations[heroNameNormalized];
    }
    
    const hero = DOTA_HEROES.find(h => 
        h.id === heroNameNormalized || 
        normalizeHeroName(h.name) === heroNameNormalized ||
        normalizeHeroName(h.localized_name) === heroNameNormalized
    );
    if (!hero) {
        console.warn(`Hero not found: ${heroName} (normalized: ${heroNameNormalized})`);
        return null;
    }
    const heroId = hero.id;
    return `/api/hero-icon/${heroId}`;
}

// Retry utility with exponential backoff
async function fetchWithRetry(url, options = {}, maxRetries = 3, baseDelay = 1000) {
    let lastError;
    for (let attempt = 0; attempt <= maxRetries; attempt++) {
        try {
            const controller = new AbortController();
            const timeoutId = setTimeout(() => controller.abort(), 30000);
            const response = await fetch(url, {
                ...options,
                signal: controller.signal
            });
            clearTimeout(timeoutId);
            
            if (!response.ok) {
                const errorText = await response.text().catch(() => '');
                if (response.status >= 500 && attempt < maxRetries) {
                    throw new Error(`Server error (${response.status}): ${errorText || response.statusText}`);
                }
                throw new Error(errorText || `HTTP ${response.status}: ${response.statusText}`);
            }
            
            return response;
        } catch (error) {
            lastError = error;
            if (attempt < maxRetries) {
                const isConnectionError = error.name === 'AbortError' || 
                                        error.message.includes('Failed to fetch') ||
                                        error.message.includes('NetworkError') ||
                                        error.message.includes('timeout');
                
                if (isConnectionError || (error.message && error.message.includes('500'))) {
                    const delay = baseDelay * Math.pow(2, attempt);
                    console.log(`Request failed (attempt ${attempt + 1}/${maxRetries + 1}), retrying in ${delay}ms...`, error.message);
                    await new Promise(resolve => setTimeout(resolve, delay));
                    continue;
                }
            }
            throw error;
        }
    }
    throw lastError;
}

document.addEventListener('DOMContentLoaded', () => {
    let currentPath = '';
    let currentReplays = [];
    
    const replayDirInput = document.getElementById('replay-dir');
    const saveConfigBtn = document.getElementById('save-config');
    const replayList = document.getElementById('replay-list');
    const refreshReplaysBtn = document.getElementById('refresh-replays');
    const sortDateBtn = document.getElementById('sort-date');
    const selectAllBtn = document.getElementById('select-all');
    const deselectAllBtn = document.getElementById('deselect-all');
    const selectLastBtn = document.getElementById('select-last');
    const selectLastCountInput = document.getElementById('select-last-count');
    const deleteSelectedBtn = document.getElementById('delete-selected');
    const startParseBtn = document.getElementById('start-parse');
    const steamIdInput = document.getElementById('steam-id');
    const playerSelect = document.getElementById('player-select');
    const playerSelectSpinner = document.getElementById('player-select-spinner');
    const steamIdGroup = document.getElementById('steam-id-group');
    const playerSelectionGroup = document.getElementById('player-selection-group');
    const progressSection = document.getElementById('progress-section');
    const progressBar = document.getElementById('progress-bar');
    const progressText = document.getElementById('progress-text');
    const resultsSection = document.getElementById('results-section');
    const graphsSection = document.getElementById('graphs-section');
    const matchSelectorContainer = document.getElementById('match-selector-container');
    const matchSelector = document.getElementById('match-selector');
    const playerSelectorContainer = document.getElementById('player-selector-container');
    const playerSelectorDisplay = document.getElementById('player-selector-display');
    const playerSelectorDropdown = document.getElementById('player-selector-dropdown');
    const timelineGraphContainer = document.getElementById('timeline-graph-container');
    const timelineGraphCanvas = document.getElementById('timeline-graph');
    let playerSelectorOptions = [];
    const historySteamIdInput = document.getElementById('history-steam-id');
    const historyLimitInput = document.getElementById('history-limit');
    const fetchHistoryBtn = document.getElementById('fetch-history');
    const historyResults = document.getElementById('history-results');

    function updatePlayerSelection() {
        const selectedCheckboxes = document.querySelectorAll('.replay-item input[type="checkbox"]:checked');
        const selectedCount = selectedCheckboxes.length;
        
        if (selectedCount > 1) {
            steamIdGroup.classList.remove('hidden');
            playerSelectionGroup.classList.add('hidden');
            playerSelect.innerHTML = '<option value="-1">Select a player...</option>';
        } else {
            steamIdGroup.classList.add('hidden');
            playerSelectionGroup.classList.add('hidden');
        }
    }
    
    async function loadPlayerInfo(matchId) {
        playerSelect.innerHTML = '<option value="-1">Loading players...</option>';
        playerSelect.disabled = true;
        playerSelectSpinner.classList.remove('hidden');
        
        const clearLoading = () => {
            playerSelectSpinner.classList.add('hidden');
        };
        
        try {
            const res = await fetch('/api/player-info', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    matchId: matchId,
                    profileName: getSelectedProfileName()
                })
            });
            
            if (!res.ok) {
                throw new Error(await res.text());
            }
            
            const players = await res.json();
            playerSelect.innerHTML = '<option value="-1">Select a player...</option>';
            
            players.forEach(player => {
                const option = document.createElement('option');
                option.value = player.Slot;
                const teamName = player.Team === 2 ? 'Radiant' : player.Team === 3 ? 'Dire' : 'Unknown';
                const playerName = player.Name || '(Empty)';
                const heroName = player.Hero || '';
                let displayText = `Slot ${player.Slot} [${teamName}] - ${playerName}`;
                if (heroName) {
                    displayText += ` (${heroName})`;
                }
                option.textContent = displayText;
                option.className = player.Team === 2 ? 'radiant-option' : player.Team === 3 ? 'dire-option' : '';
                playerSelect.appendChild(option);
            });
            
            playerSelect.disabled = false;
            clearLoading();
        } catch (err) {
            playerSelect.innerHTML = '<option value="-1">Error loading players</option>';
            clearLoading();
            console.error('Error loading player info:', err);
        }
    }

    // Profile Elements
    const profileSelect = document.getElementById('profile-select');
    const newProfileName = document.getElementById('new-profile-name');
    const newProfileId = document.getElementById('new-profile-id');
    const addProfileBtn = document.getElementById('add-profile-btn');
    const deleteProfileBtn = document.getElementById('delete-profile-btn');

    let profiles = [];
    let isInitialLoad = true;

    function loadProfiles() {
        const savedProfiles = localStorage.getItem('steamProfiles');
        if (savedProfiles) {
            try {
                profiles = JSON.parse(savedProfiles);
            } catch (e) {
                console.error('Failed to parse profiles', e);
                profiles = [];
            }
        }
        renderProfileSelect(true);
        isInitialLoad = false;
    }

    function saveProfiles() {
        localStorage.setItem('steamProfiles', JSON.stringify(profiles));
        renderProfileSelect(false);
    }

    function renderProfileSelect(shouldLoadReplays) {
        const savedIndex = localStorage.getItem('selectedProfileIndex');
        const currentVal = savedIndex !== null && profiles[savedIndex] ? savedIndex : profileSelect.value;
        
        profileSelect.innerHTML = '<option value="">Select a profile...</option>';
        profiles.forEach((p, index) => {
            const option = document.createElement('option');
            option.value = index;
            option.textContent = `${p.name} (${p.id})`;
            profileSelect.appendChild(option);
        });

        if (currentVal && profiles[currentVal]) {
            profileSelect.value = currentVal;
            localStorage.setItem('selectedProfileIndex', currentVal);
            deleteProfileBtn.style.visibility = 'visible';
            if (shouldLoadReplays) {
                const profileName = profiles[currentVal].name;
                browseDirectory('', profileName);
            }
        } else {
            profileSelect.value = "";
            localStorage.removeItem('selectedProfileIndex');
            deleteProfileBtn.style.visibility = 'hidden';
        }
    }

    addProfileBtn.addEventListener('click', () => {
        const name = newProfileName.value.trim();
        const id = newProfileId.value.trim();

        if (!name || !id) {
            alert('Please enter both a name and a Steam ID');
            return;
        }

        profiles.push({ name, id });
        saveProfiles();
        
        // Clear inputs
        newProfileName.value = '';
        newProfileId.value = '';
        
        // Select the new profile automatically
        profileSelect.value = profiles.length - 1;
        profileSelect.dispatchEvent(new Event('change'));
    });

    deleteProfileBtn.addEventListener('click', () => {
        const index = profileSelect.value;
        if (index === "") return;

        if (confirm(`Delete profile "${profiles[index].name}"?`)) {
            profiles.splice(index, 1);
            saveProfiles();
            profileSelect.value = "";
            deleteProfileBtn.style.visibility = 'hidden';
            
            // Optionally clear the filled fields or leave them?
            // Leaving them is safer/less annoying.
        }
    });

    profileSelect.addEventListener('change', () => {
        console.log('Profile select changed, value:', profileSelect.value);
        const index = profileSelect.value;
        let profileName = "";
        if (index !== "" && profiles[index]) {
            const p = profiles[index];
            profileName = p.name;
            localStorage.setItem('selectedProfileIndex', index);
            console.log('Selected profile:', profileName);
            if (historySteamIdInput) {
                historySteamIdInput.value = p.id;
                historySteamIdInput.dispatchEvent(new Event('input')); 
            }
            if (steamIdInput) {
                steamIdInput.value = p.id;
                steamIdInput.dispatchEvent(new Event('input'));
                if (typeof syncSteamIdSlotId === 'function') {
                    syncSteamIdSlotId();
                }
            }
            const fatalSteamIdInput = document.getElementById('fatal-steam-id');
            if (fatalSteamIdInput) {
                fatalSteamIdInput.value = p.id;
            }
            deleteProfileBtn.style.visibility = 'visible';
        } else {
            console.log('No profile selected');
            localStorage.removeItem('selectedProfileIndex');
            deleteProfileBtn.style.visibility = 'hidden';
            const fatalSteamIdInput = document.getElementById('fatal-steam-id');
            if (fatalSteamIdInput) {
                fatalSteamIdInput.value = '';
            }
        }
        console.log('Calling browseDirectory with profileName:', profileName);
        browseDirectory('', profileName);
    });

    function getSelectedProfileName() {
        const index = profileSelect.value;
        if (index !== "" && profiles[index]) {
            return profiles[index].name;
        }
        return "";
    }

    function getSelectedProfile() {
        const index = profileSelect.value;
        if (index !== "" && profiles[index]) {
            return profiles[index];
        }
        return null;
    }

    // Initialize Profiles
    loadProfiles();
    
    // Steam Login Elements
    const steamUserInput = document.getElementById('steam-user');
    const steamPassInput = document.getElementById('steam-pass');
    const steamApiKeyInput = document.getElementById('steam-api-key');
    const steamCodeInput = document.getElementById('steam-code');
    const steamGuardGroup = document.getElementById('steam-guard-group');
    const steamLoginBtn = document.getElementById('steam-login-btn');
    const steamDisconnectBtn = document.getElementById('steam-disconnect-btn');
    const steamStatusText = document.getElementById('steam-status');

    // Steam Logic
    let steamPollingInterval;
    let connectingTimeoutId = null;
    let submittingTimeoutId = null;

    let lastSteamStatus = null;
    let lastSteamStatusText = null;
    
    function updateSteamUI(status, text, errorMessage) {
        const statusText = `Status: ${text}`;
        if (steamStatusText.textContent !== statusText) {
            steamStatusText.textContent = statusText;
        }
        
        // Display error message if present
        let errorDisplay = document.getElementById('steam-error-message');
        if (errorMessage && errorMessage.trim() !== '') {
            if (!errorDisplay) {
                errorDisplay = document.createElement('p');
                errorDisplay.id = 'steam-error-message';
                errorDisplay.style.color = '#ef4444';
                errorDisplay.style.marginTop = '8px';
                errorDisplay.style.marginBottom = '8px';
                errorDisplay.style.fontSize = '0.9em';
                steamStatusText.parentNode.insertBefore(errorDisplay, steamStatusText.nextSibling);
            }
            errorDisplay.textContent = errorMessage;
            errorDisplay.style.display = 'block';
        } else if (errorDisplay) {
            errorDisplay.style.display = 'none';
        }
        
        if (connectingTimeoutId) {
            clearTimeout(connectingTimeoutId);
            connectingTimeoutId = null;
        }
        if (submittingTimeoutId) {
            clearTimeout(submittingTimeoutId);
            submittingTimeoutId = null;
        }
        
        if (status === 1) {
            if (lastSteamStatus !== 1) {
                steamStatusText.style.color = '#ff9800';
                steamGuardGroup.classList.add('hidden');
                steamLoginBtn.textContent = 'Connecting...';
                steamLoginBtn.disabled = true;
                steamDisconnectBtn.style.display = 'none';
                lastSteamStatus = 1;
            }
            if (steamPollingInterval) clearInterval(steamPollingInterval);
            steamPollingInterval = setInterval(pollSteamStatus, 2000);
            
            connectingTimeoutId = setTimeout(() => {
                if (steamLoginBtn.disabled && steamLoginBtn.textContent === 'Connecting...') {
                    console.warn('Connection timeout - re-enabling button as fallback');
                    steamLoginBtn.disabled = false;
                    steamLoginBtn.textContent = 'Connect to Steam';
                    steamStatusText.textContent = 'Status: Disconnected (timeout)';
                    steamStatusText.style.color = '';
                }
            }, 20000);
        } else if (status === 2) {
            if (lastSteamStatus !== 2) {
                steamStatusText.style.color = '#ff9800';
                steamGuardGroup.classList.remove('hidden');
                if (steamLoginBtn.textContent === 'Submitting Code...') {
                    submittingTimeoutId = setTimeout(() => {
                        if (steamLoginBtn.disabled && steamLoginBtn.textContent === 'Submitting Code...') {
                            console.warn('Submit code timeout - re-enabling button as fallback');
                            steamLoginBtn.disabled = false;
                            steamLoginBtn.textContent = 'Submit Code';
                        }
                    }, 15000);
                } else {
                    steamLoginBtn.textContent = 'Submit Code';
                    steamLoginBtn.disabled = false;
                }
                steamDisconnectBtn.style.display = 'none';
                lastSteamStatus = 2;
            }
            if (steamPollingInterval) clearInterval(steamPollingInterval);
            steamPollingInterval = setInterval(pollSteamStatus, 2000);
        } else if (status === 3 || status === 4) {
             if (lastSteamStatus !== status) {
                 steamStatusText.style.color = '#4caf50';
                 steamLoginBtn.textContent = status === 4 ? 'Connected' : 'Connecting...';
                 steamLoginBtn.disabled = true;
                 steamGuardGroup.classList.add('hidden');
                 steamDisconnectBtn.style.display = 'inline-block';
                 lastSteamStatus = status;
                 if (status === 4) {
                     fetchConductScorecard();
                 }
             }
             if (steamPollingInterval) clearInterval(steamPollingInterval);
             steamPollingInterval = setInterval(pollSteamStatus, 10000);
        } else if (status === 5) { // Rate Limited
             if (lastSteamStatus !== 5) {
                 steamStatusText.style.color = '#f44336';
                 steamStatusText.textContent = 'Status: Rate Limited (Wait 24h)';
                 steamLoginBtn.textContent = 'Rate Limited';
                 steamLoginBtn.disabled = true;
                 steamGuardGroup.classList.add('hidden');
                 steamDisconnectBtn.style.display = 'inline-block';
                 lastSteamStatus = 5;
             }
             if (steamPollingInterval) clearInterval(steamPollingInterval);
             steamPollingInterval = setInterval(pollSteamStatus, 60000);
        } else {
            if (lastSteamStatus !== status) {
                steamStatusText.style.color = '';
                steamGuardGroup.classList.add('hidden');
                steamCodeInput.value = '';
                steamLoginBtn.textContent = 'Connect to Steam';
                steamLoginBtn.disabled = false;
                steamDisconnectBtn.style.display = 'none';
                document.getElementById('conduct-scorecard-section').classList.add('hidden');
                lastSteamStatus = status;
            }
            if (steamPollingInterval) clearInterval(steamPollingInterval);
            steamPollingInterval = setInterval(pollSteamStatus, 2000);
        }
    }

    let lastConductScorecardData = null;
    let isFetchingConductScorecard = false;
    
    function fetchConductScorecard() {
        const scorecardSection = document.getElementById('conduct-scorecard-section');
        const scorecardContent = document.getElementById('conduct-scorecard-content');
        
        if (!scorecardSection || !scorecardContent) return;
        if (isFetchingConductScorecard) return;
        
        isFetchingConductScorecard = true;
        const wasHidden = scorecardSection.classList.contains('hidden');
        if (wasHidden) {
            scorecardSection.classList.remove('hidden');
        }
        if (!scorecardContent.querySelector('.conduct-scorecard-grid')) {
            scorecardContent.innerHTML = '<div class="loading">Loading conduct scorecard...</div>';
        }
        
        fetchWithRetry('/api/steam/conduct-scorecard', {}, 2, 500)
            .then(res => {
                if (!res.ok) {
                    throw new Error(`HTTP ${res.status}: ${res.statusText}`);
                }
                return res.json();
            })
            .then(data => {
                const dataKey = JSON.stringify(data);
                if (dataKey !== lastConductScorecardData) {
                    displayConductScorecard(data);
                    lastConductScorecardData = dataKey;
                }
                isFetchingConductScorecard = false;
            })
            .catch(err => {
                console.error('Error fetching conduct scorecard:', err);
                if (!scorecardContent.querySelector('.conduct-scorecard-grid')) {
                    scorecardContent.innerHTML = `<div style="color: var(--danger-color);">Error loading conduct scorecard: ${err.message}</div>`;
                }
                isFetchingConductScorecard = false;
            });
    }

    function displayConductScorecard(data) {
        const scorecardContent = document.getElementById('conduct-scorecard-content');
        if (!scorecardContent) return;
        
        const formatDate = (timestamp) => {
            if (!timestamp || timestamp === 0) return 'N/A';
            const date = new Date(timestamp * 1000);
            return date.toLocaleString('en-US', {
                year: 'numeric',
                month: 'short',
                day: 'numeric',
                hour: '2-digit',
                minute: '2-digit'
            });
        };
        
        const rawBehaviorScore = data.raw_behavior_score || 0;
        const oldRawBehaviorScore = data.old_raw_behavior_score || 0;
        const behaviorScoreChange = rawBehaviorScore - oldRawBehaviorScore;
        const behaviorScoreChangeClass = behaviorScoreChange >= 0 ? 'positive' : 'negative';
        const behaviorScoreChangeSymbol = behaviorScoreChange >= 0 ? '↑' : '↓';
        
        let html = '<div class="conduct-scorecard-grid">';
        
        html += `
            <div class="conduct-stat-card">
                <div class="conduct-stat-label">Raw Behavior Score</div>
                <div class="conduct-stat-value">${rawBehaviorScore.toLocaleString()}</div>
                ${oldRawBehaviorScore > 0 ? `
                    <div class="conduct-stat-change ${behaviorScoreChangeClass}">
                        ${behaviorScoreChangeSymbol} ${Math.abs(behaviorScoreChange).toLocaleString()} 
                        (was ${oldRawBehaviorScore.toLocaleString()})
                    </div>
                ` : ''}
            </div>
        `;
        
        html += `
            <div class="conduct-stat-card">
                <div class="conduct-stat-label">Commend Count</div>
                <div class="conduct-stat-value">${data.commend_count || 0}</div>
            </div>
        `;
        
        html += `
            <div class="conduct-stat-card">
                <div class="conduct-stat-label">Matches in Report</div>
                <div class="conduct-stat-value">${data.matches_in_report || 0}</div>
                <div class="conduct-date-display">15-game period</div>
            </div>
        `;
        
        html += `
            <div class="conduct-stat-card">
                <div class="conduct-stat-label">Matches Clean</div>
                <div class="conduct-stat-value">${data.matches_clean || 0}</div>
            </div>
        `;
        
        html += `
            <div class="conduct-stat-card">
                <div class="conduct-stat-label">Matches Reported</div>
                <div class="conduct-stat-value" style="color: ${(data.matches_reported || 0) > 0 ? 'var(--warning-color)' : 'var(--text-primary)'}">${data.matches_reported || 0}</div>
            </div>
        `;
        
        html += `
            <div class="conduct-stat-card">
                <div class="conduct-stat-label">Matches Abandoned</div>
                <div class="conduct-stat-value" style="color: ${(data.matches_abandoned || 0) > 0 ? 'var(--danger-color)' : 'var(--text-primary)'}">${data.matches_abandoned || 0}</div>
            </div>
        `;
        
        html += `
            <div class="conduct-stat-card">
                <div class="conduct-stat-label">Match Date of Last Chunk</div>
                <div class="conduct-stat-value" style="font-size: 1rem;">${formatDate(data.date)}</div>
            </div>
        `;
        
        html += `
            <div class="conduct-stat-card">
                <div class="conduct-stat-label">Match ID</div>
                <div class="conduct-match-id">${data.match_id || 'N/A'}</div>
            </div>
        `;
        
        html += '</div>';
        
        const reasons = data.reasons || 0;
        const reasonNames = [];
        if (reasons & 1) reasonNames.push('Cheating');
        if (reasons & 2) reasonNames.push('Feeding');
        if (reasons & 4) reasonNames.push('Griefing');
        if (reasons & 8) reasonNames.push('Suspicious');
        if (reasons & 16) reasonNames.push('AbilityAbuse');
        if (reasons === 0) reasonNames.push('Unknown');
        
        const reasonsText = reasonNames.length > 0 ? ` (${reasonNames.join(', ')})` : '';
        
        html += `
            <div class="conduct-reasons-card">
                <h3>Reasons</h3>
                <div class="conduct-reasons-value">${reasons}${reasonsText}</div>
                <div class="conduct-reasons-note">Report reasons from the conduct scorecard.</div>
            </div>
        `;
        
        if (data.match_id && data.account_id) {
            html += `
                <div style="margin-top: var(--spacing-md); display: flex; justify-content: center; gap: var(--spacing-md); flex-wrap: wrap;">
                    <button id="validate-report-card-btn" class="btn primary-btn" 
                            data-match-id="${data.match_id}" 
                            data-account-id="${data.account_id}">
                        Validate Report Card
                    </button>
                    <button id="validate-report-card-current-btn" class="btn secondary-btn" 
                            data-match-id="${data.match_id}" 
                            data-account-id="${data.account_id}">
                        Current
                    </button>
                </div>
                <div id="validate-report-card-status" style="margin-top: var(--spacing-sm); text-align: center; color: var(--text-secondary); font-size: 0.9rem;"></div>
            `;
        }
        
        html += `
            <div class="conduct-theories-card">
                <h3>Why counts might not match analysis:</h3>
                <ul class="conduct-theories-list">
                    <li>Most reports may be ignored if players abuse the reporting system</li>
                    <li>There may be a hidden limit to reports that we don't realize</li>
                    <li>Report counts may refresh on a certain day of the week</li>
                    <li>The overwatch report checkbox - when reporting someone, there's a checkbox if you have an 'overwatch report' available</li>
                </ul>
            </div>
        `;
        
        scorecardContent.innerHTML = html;
        
        const validateBtn = document.getElementById('validate-report-card-btn');
        if (validateBtn && !validateBtn.hasAttribute('data-listener-attached')) {
            validateBtn.setAttribute('data-listener-attached', 'true');
            validateBtn.addEventListener('click', () => {
                const matchId = validateBtn.getAttribute('data-match-id');
                const accountId = validateBtn.getAttribute('data-account-id');
                validateReportCard(parseInt(matchId), parseInt(accountId), false);
            });
        }
        
        const validateCurrentBtn = document.getElementById('validate-report-card-current-btn');
        if (validateCurrentBtn && !validateCurrentBtn.hasAttribute('data-listener-attached')) {
            validateCurrentBtn.setAttribute('data-listener-attached', 'true');
            validateCurrentBtn.addEventListener('click', () => {
                const matchId = validateCurrentBtn.getAttribute('data-match-id');
                const accountId = validateCurrentBtn.getAttribute('data-account-id');
                validateReportCard(parseInt(matchId), parseInt(accountId), true);
            });
        }
    }

    let validateReportCardInProgress = false;

    function validateReportCard(matchId, accountId, isCurrent) {
        const statusDiv = document.getElementById('validate-report-card-status');
        const btn = isCurrent ? document.getElementById('validate-report-card-current-btn') : document.getElementById('validate-report-card-btn');
        const otherBtn = isCurrent ? document.getElementById('validate-report-card-btn') : document.getElementById('validate-report-card-current-btn');
        
        if (!statusDiv || !btn) return;
        
        if (validateReportCardInProgress) {
            statusDiv.textContent = 'Another validation is already in progress. Please wait...';
            statusDiv.style.color = 'var(--warning-color)';
            return;
        }
        
        validateReportCardInProgress = true;
        btn.disabled = true;
        if (otherBtn) otherBtn.disabled = true;
        btn.textContent = isCurrent ? 'Fetching Current...' : 'Validating...';
        statusDiv.textContent = isCurrent ? 'Fetching games after match ID...' : 'Fetching match history and downloading replays (this may take several minutes)...';
        statusDiv.style.color = 'var(--text-secondary)';
        
        const endpoint = isCurrent ? '/api/steam/validate-report-card-current' : '/api/steam/validate-report-card';
        
        fetch(endpoint, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                matchId: matchId,
                accountId: accountId
            })
        })
            .then(res => {
                if (!res.ok) {
                    return res.text().then(text => {
                        throw new Error(text || `HTTP ${res.status}`);
                    });
                }
                return res.json();
            })
            .then(data => {
                const total = data.total || 0;
                const downloaded = data.downloaded?.length || 0;
                const skipped = data.skipped?.length || 0;
                const errors = data.errors?.length || 0;
                
                let message = '';
                if (data.message) {
                    message = data.message;
                } else {
                    message = `Downloaded: ${downloaded}, Skipped: ${skipped}`;
                    if (errors > 0) {
                        message += `, Errors: ${errors}`;
                    }
                    message += ` (Total: ${total})`;
                }
                
                if (data.directory) {
                    message += `\nSaved to: ${data.directory}`;
                }
                
                statusDiv.textContent = message;
                statusDiv.style.color = errors > 0 ? 'var(--warning-color)' : (data.message && data.message.includes('not downloading') ? 'var(--text-secondary)' : 'var(--success-color)');
                validateReportCardInProgress = false;
                btn.disabled = false;
                btn.textContent = isCurrent ? 'Current' : 'Validate Report Card';
                if (otherBtn) otherBtn.disabled = false;
            })
            .catch(err => {
                console.error('Error validating report card:', err);
                if (err.name === 'AbortError' || err.message.includes('aborted')) {
                    statusDiv.textContent = 'Request was cancelled. The download may still be in progress on the server.';
                    statusDiv.style.color = 'var(--warning-color)';
                } else {
                    statusDiv.textContent = `Error: ${err.message}`;
                    statusDiv.style.color = 'var(--danger-color)';
                }
                validateReportCardInProgress = false;
                btn.disabled = false;
                btn.textContent = isCurrent ? 'Current' : 'Validate Report Card';
                if (otherBtn) otherBtn.disabled = false;
            });
    }

    function pollSteamStatus() {
        fetchWithRetry('/api/steam/status', {}, 2, 500)
            .then(res => res.json())
            .then(data => {
                updateSteamUI(data.status, data.statusText, data.errorMessage);
            })
            .catch(err => {
                console.error('Steam status poll failed:', err);
                // Don't show error for polling failures, just log
            });
    }

    // Start polling on load
    pollSteamStatus();
    steamPollingInterval = setInterval(pollSteamStatus, 2000);

    if (steamLoginBtn) {
        steamLoginBtn.addEventListener('click', () => {
        const username = steamUserInput.value.trim();
        const password = steamPassInput.value.trim();
        const isSubmittingCode = steamLoginBtn.textContent === 'Submit Code';
        const code = isSubmittingCode ? steamCodeInput.value.trim() : '';

        if (!username || !password) {
            alert('Please enter both username and password');
            return;
        }

        if (isSubmittingCode && !code) {
            alert('Please enter your Steam Guard code');
            return;
        }

        const originalText = steamLoginBtn.textContent;
        steamLoginBtn.disabled = true;
        steamLoginBtn.textContent = code ? 'Submitting Code...' : 'Connecting...';

        const controller = new AbortController();
        const timeoutId = setTimeout(() => controller.abort(), 35000);
        
        fetchWithRetry('/api/steam/login', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username, password, code }),
            signal: controller.signal
        }, 2, 2000)
        .then(async res => {
            clearTimeout(timeoutId);
            const contentType = res.headers.get('content-type');
            if (!res.ok) {
                let errorText = '';
                if (contentType && contentType.includes('application/json')) {
                    const data = await res.json();
                    errorText = data.error || data.message || `HTTP ${res.status}`;
                    // If we got status info, update UI
                    if (data.status !== undefined) {
                        pollSteamStatus();
                        if (data.status === 2) {
                            steamLoginBtn.disabled = false;
                            steamLoginBtn.textContent = 'Submit Code';
                        }
                    }
                } else {
                    errorText = await res.text();
                }
                throw new Error(errorText || `Server error (${res.status})`);
            }
            return res.json();
        })
        .then(data => {
             pollSteamStatus();
             setTimeout(() => {
                 if (steamLoginBtn.textContent.includes('Submitting') || steamLoginBtn.textContent.includes('Connecting')) {
                     if (data.status === 2) {
                         steamLoginBtn.disabled = false;
                         steamLoginBtn.textContent = 'Submit Code';
                     } else if (data.status !== 3 && data.status !== 4) {
                         steamLoginBtn.disabled = false;
                         steamLoginBtn.textContent = originalText;
                     }
                 }
             }, 2000);
        })
        .catch(err => {
            clearTimeout(timeoutId);
            console.error('Steam login error:', err);
            // Try to get current status to update UI
            pollSteamStatus();
            const errorMsg = err.name === 'AbortError' ? 'Request timed out. Please check your connection and try again.' : (err.message || 'Failed to send login request');
            if (err.name !== 'AbortError') {
                alert('Failed to send login request: ' + errorMsg);
            } else {
                // For timeout, just update UI without alert
                console.warn('Login request timed out');
            }
            steamLoginBtn.disabled = false;
            steamLoginBtn.textContent = originalText;
            // Clear error message display if present
            const errorDisplay = document.getElementById('steam-error-message');
            if (errorDisplay) {
                errorDisplay.style.display = 'none';
            }
        });
    });
    }

    // Disconnect button handler
    if (steamDisconnectBtn) {
        steamDisconnectBtn.addEventListener('click', () => {
            if (!confirm('Are you sure you want to disconnect from Steam?')) {
            return;
        }

        const originalText = steamDisconnectBtn.textContent;
        steamDisconnectBtn.disabled = true;
        steamDisconnectBtn.textContent = 'Disconnecting...';

        fetchWithRetry('/api/steam/disconnect', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' }
        }, 2, 1000)
        .then(async res => {
            if (!res.ok) {
                const text = await res.text();
                throw new Error(`Server error (${res.status}): ${text}`);
            }
            return res.json();
        })
        .then(data => {
            console.log('Disconnect successful:', data);
            steamCodeInput.value = '';
            pollSteamStatus();
            steamDisconnectBtn.disabled = false;
            steamDisconnectBtn.textContent = originalText;
        })
        .catch(err => {
            console.error('Steam disconnect error:', err);
            alert('Failed to disconnect: ' + err.message);
            steamDisconnectBtn.disabled = false;
            steamDisconnectBtn.textContent = originalText;
        });
        });
    }
    
    // Load from localStorage
    function loadFromStorage() {
        const savedReplayDir = localStorage.getItem('replayDir');
        const savedSteamUser = localStorage.getItem('steamUser');
        const savedSteamPass = localStorage.getItem('steamPass');
        const savedSteamId = localStorage.getItem('steamId');
        const savedSteamApiKey = localStorage.getItem('steamApiKey');
        
        if (savedReplayDir) replayDirInput.value = savedReplayDir;
        if (savedSteamUser) steamUserInput.value = savedSteamUser;
        if (savedSteamPass) steamPassInput.value = savedSteamPass;
        if (savedSteamId) steamIdInput.value = savedSteamId;
        if (savedSteamApiKey) steamApiKeyInput.value = savedSteamApiKey;
    }

    // Save to localStorage
    function saveToStorage() {
        if (replayDirInput.value) localStorage.setItem('replayDir', replayDirInput.value);
        if (steamUserInput.value) localStorage.setItem('steamUser', steamUserInput.value);
        if (steamPassInput.value) localStorage.setItem('steamPass', steamPassInput.value);
        if (steamIdInput.value) localStorage.setItem('steamId', steamIdInput.value);
        if (steamApiKeyInput.value) localStorage.setItem('steamApiKey', steamApiKeyInput.value);
    }

    // Load from localStorage first (before server config)
    loadFromStorage();

    // Load Config from server (will override localStorage if present)
    fetchWithRetry('/api/config', {}, 2, 1000)
        .then(res => res.json())
        .then(config => {
            replayDirInput.value = config.replayDir || '';
            
            // Only update steamApiKey if server has a non-empty value
            // This preserves localStorage value if server returns empty string
            if (config.steamApiKey && config.steamApiKey.trim() !== '') {
                steamApiKeyInput.value = config.steamApiKey;
                localStorage.setItem('steamApiKey', config.steamApiKey);
            }
            // If server returns empty/undefined, keep what was loaded from localStorage

            if (config.steamUser && config.steamUser.trim() !== '') {
                steamUserInput.value = config.steamUser;
                localStorage.setItem('steamUser', config.steamUser);
            }
            
            if (config.steamPass && config.steamPass.trim() !== '') {
                steamPassInput.value = config.steamPass;
                localStorage.setItem('steamPass', config.steamPass);
            }
            browseDirectory(currentPath);
        })
        .catch(() => {
        });

    // Auto-save config fields when they change (with debounce)
    let saveTokenTimeout;
    const autoSave = () => {
        const replayDir = replayDirInput.value.trim();
        const steamUser = steamUserInput.value.trim();
        const steamPass = steamPassInput.value.trim();
        const steamApiKey = steamApiKeyInput.value.trim();
        
        if (replayDir) localStorage.setItem('replayDir', replayDir);
        if (steamUser) localStorage.setItem('steamUser', steamUser);
        if (steamPass) localStorage.setItem('steamPass', steamPass);
        if (steamApiKey) localStorage.setItem('steamApiKey', steamApiKey);
        
        // Auto-save to server after user stops typing (1 second delay)
        clearTimeout(saveTokenTimeout);
        saveTokenTimeout = setTimeout(() => {
            const payload = {};
            if (replayDir) payload.replayDir = replayDir;
            if (steamUser) payload.steamUser = steamUser;
            if (steamPass) payload.steamPass = steamPass;
            if (steamApiKey) payload.steamApiKey = steamApiKey;

            if (Object.keys(payload).length > 0) {
                fetchWithRetry('/api/config', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(payload)
                }, 2, 1000).catch(err => console.error('Auto-save failed:', err));
            }
        }, 1000);
    };

    replayDirInput.addEventListener('input', autoSave);
    steamUserInput.addEventListener('input', autoSave);
    steamPassInput.addEventListener('input', autoSave);
    steamApiKeyInput.addEventListener('input', autoSave);
    
    steamIdInput.addEventListener('input', () => {
        if (steamIdInput.value) localStorage.setItem('steamId', steamIdInput.value);
    });
    
    replayList.addEventListener('change', (e) => {
        if (e.target.type === 'checkbox' && e.target.closest('.replay-item')) {
            updatePlayerSelection();
        }
    });

    // Save Config
    saveConfigBtn.addEventListener('click', () => {
        const newDir = replayDirInput.value.trim();
        const newApiKey = steamApiKeyInput.value.trim();
        const newSteamUser = steamUserInput.value.trim();
        const newSteamPass = steamPassInput.value.trim();

        saveToStorage();
        
        const payload = {};
        if (newDir) payload.replayDir = newDir;
        if (newApiKey) payload.steamApiKey = newApiKey;
        if (newSteamUser) payload.steamUser = newSteamUser;
        if (newSteamPass) payload.steamPass = newSteamPass;
        
        fetchWithRetry('/api/config', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload)
        }, 2, 1000)
            .then(res => res.json())
            .then(config => {
                if (config.replayDir) {
                    localStorage.setItem('replayDir', config.replayDir);
                    replayDirInput.value = config.replayDir;
                }
                if (config.steamUser) {
                    localStorage.setItem('steamUser', config.steamUser);
                    steamUserInput.value = config.steamUser;
                }
                if (config.steamPass) {
                    localStorage.setItem('steamPass', config.steamPass);
                    steamPassInput.value = config.steamPass;
                }
                if (config.steamApiKey) {
                    localStorage.setItem('steamApiKey', config.steamApiKey);
                    steamApiKeyInput.value = config.steamApiKey;
                }
                alert('Configuration saved!');
                browseDirectory(currentPath);
            })
            .catch(err => {
                console.error('Error saving config:', err);
                alert('Error saving config: ' + err);
            });
    });

    const browseBackBtn = document.getElementById('browse-back');
    
    function browseDirectory(path = '', profileNameOverride, silent = false) {
        if (!replayList) {
            console.error('replayList element not found');
            return;
        }
        if (!silent) {
            replayList.innerHTML = '<p class="loading">Loading...</p>';
        }
        const profileName = profileNameOverride !== undefined ? profileNameOverride : getSelectedProfileName();
        const url = '/api/browse?t=' + Date.now() + (profileName ? '&profile=' + encodeURIComponent(profileName) : '') + (path ? '&path=' + encodeURIComponent(path) : '');
        
        fetch(url)
            .then(res => {
                if (!res.ok) {
                    throw new Error(`HTTP ${res.status}: ${res.statusText}`);
                }
                return res.json();
            })
            .then(items => {
                const pathChanged = currentPath !== path;
                currentPath = path;
                browseBackBtn.style.display = path ? 'inline-block' : 'none';
                renderBrowseList(items, silent && !pathChanged);
                if (window.updateHistoryStatus && !silent) {
                    window.updateHistoryStatus();
                }
            })
            .catch(err => {
                console.error('Error browsing directory:', err);
                if (!silent) {
                    replayList.innerHTML = `<p style="color:red">Error browsing directory: ${err}</p>`;
                }
            });
    }

    setInterval(() => {
        if (document.visibilityState === 'visible') {
            browseDirectory(currentPath, undefined, true);
        }
    }, 120000);
    
    function renderBrowseList(items, preserveState = false) {
        if (!items || items.length === 0) {
            if (!preserveState) {
                replayList.innerHTML = '<p>Directory is empty.</p>';
            }
            return;
        }

        if (preserveState) {
            const scrollTop = replayList.scrollTop;
            const checkedPaths = new Set(Array.from(document.querySelectorAll('.replay-item input[type="checkbox"]:checked')).map(cb => cb.value));
            const existingItems = new Map();
            Array.from(replayList.children).forEach(child => {
                const checkbox = child.querySelector('input[type="checkbox"]');
                if (checkbox) {
                    existingItems.set(checkbox.value, { checked: checkbox.checked, element: child });
                } else {
                    const label = child.querySelector('label');
                    if (label) {
                        const dirName = label.textContent.trim().replace('/', '');
                        existingItems.set(`dir:${dirName}`, { element: child });
                    }
                }
            });

            const newItems = new Set();
            items.forEach(item => {
                if (item.isFile && item.name.endsWith('.dem')) {
                    const fullPath = currentPath ? `${currentPath}/${item.name}` : item.name;
                    newItems.add(fullPath);
                } else if (item.isDir) {
                    newItems.add(`dir:${item.name}`);
                }
            });

            const fragment = document.createDocumentFragment();
            let hasChanges = false;

            existingItems.forEach((data, path) => {
                if (!newItems.has(path)) {
                    data.element.remove();
                    hasChanges = true;
                }
            });

            items.forEach(item => {
                if (item.isFile && item.name.endsWith('.dem')) {
                    const matchId = item.name.replace('.dem', '');
                    const fullPath = currentPath ? `${currentPath}/${item.name}` : item.name;
                    const escapedPath = fullPath.replace(/"/g, '&quot;');
                    const existing = existingItems.get(fullPath);
                    
                    if (!existing) {
                        const div = document.createElement('div');
                        div.className = 'replay-item';
                        const date = item.date ? new Date(item.date).toLocaleString() : '';
                        const dateDisplay = date ? ` <span class="date-display">(${date})</span>` : '';
                        div.innerHTML = `
                            <input type="checkbox" value="${escapedPath}" id="replay-${matchId}" data-match-id="${matchId}" ${checkedPaths.has(fullPath) ? 'checked' : ''}>
                            <label for="replay-${matchId}" style="cursor: pointer;">${item.name}${dateDisplay}</label>
                        `;
                        fragment.appendChild(div);
                        hasChanges = true;
                    } else {
                        const checkbox = existing.element.querySelector('input[type="checkbox"]');
                        if (checkbox && checkbox.checked !== checkedPaths.has(fullPath)) {
                            checkbox.checked = checkedPaths.has(fullPath);
                        }
                    }
                } else if (item.isDir) {
                    const existing = existingItems.get(`dir:${item.name}`);
                    if (!existing) {
                        const div = document.createElement('div');
                        div.className = 'replay-item';
                        const escapedPath = item.path.replace(/'/g, "\\'");
                        div.innerHTML = `
                            <span style="margin-right: 8px;">📁</span>
                            <label style="cursor: pointer; flex: 1;" onclick="window.browseDirectory('${escapedPath}')">${item.name}/</label>
                        `;
                        fragment.appendChild(div);
                        hasChanges = true;
                    }
                }
            });

            if (hasChanges && fragment.hasChildNodes()) {
                replayList.appendChild(fragment);
            }
            replayList.scrollTop = scrollTop;
        } else {
            replayList.innerHTML = '';
            items.forEach(item => {
                const div = document.createElement('div');
                div.className = 'replay-item';
                
                if (item.isDir) {
                    const escapedPath = item.path.replace(/'/g, "\\'");
                    div.innerHTML = `
                        <span style="margin-right: 8px;">📁</span>
                        <label style="cursor: pointer; flex: 1;" onclick="window.browseDirectory('${escapedPath}')">${item.name}/</label>
                    `;
                } else if (item.isFile && item.name.endsWith('.dem')) {
                    const matchId = item.name.replace('.dem', '');
                    const fullPath = currentPath ? `${currentPath}/${item.name}` : item.name;
                    const escapedPath = fullPath.replace(/"/g, '&quot;');
                    const date = item.date ? new Date(item.date).toLocaleString() : '';
                    const dateDisplay = date ? ` <span class="date-display">(${date})</span>` : '';
                    div.innerHTML = `
                        <input type="checkbox" value="${escapedPath}" id="replay-${matchId}" data-match-id="${matchId}">
                        <label for="replay-${matchId}" style="cursor: pointer;">${item.name}${dateDisplay}</label>
                    `;
                } else {
                    return;
                }
                
                replayList.appendChild(div);
            });
        }
    }

    if (browseBackBtn) {
        browseBackBtn.addEventListener('click', () => {
            const parts = currentPath.split('/').filter(p => p);
            parts.pop();
            const newPath = parts.join('/');
            browseDirectory(newPath);
        });
    }
    
    window.browseDirectory = browseDirectory;
    window.loadReplays = () => browseDirectory('');

    function deleteSelectedReplays() {
        const checkboxes = document.querySelectorAll('.replay-item input[type="checkbox"]:checked');
        const selectedPaths = Array.from(checkboxes).map(cb => cb.value);
        
        if (selectedPaths.length === 0) {
            alert('Please select at least one replay to delete.');
            return;
        }
        
        const fileNames = selectedPaths.map(path => path.split('/').pop()).join(', ');
        if (!confirm(`Are you sure you want to delete ${selectedPaths.length} replay file(s)?\n\n${fileNames}\n\nThis cannot be undone.`)) {
            return;
        }
        
        const originalText = deleteSelectedBtn.textContent;
        deleteSelectedBtn.disabled = true;
        deleteSelectedBtn.textContent = 'Deleting...';
        
        const deletePromises = selectedPaths.map(filePath => {
            const matchId = filePath.split('/').pop().replace('.dem', '');
            return fetch('/api/delete', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ 
                    matchId: matchId,
                    filePath: filePath,
                    profileName: getSelectedProfileName()
                })
            }).then(async res => {
                if (!res.ok) {
                    const errorText = await res.text();
                    throw new Error(`Failed to delete ${filePath}: ${errorText || res.statusText}`);
                }
                return res.json();
            });
        });
        
        Promise.allSettled(deletePromises)
            .then(results => {
                const successful = results.filter(r => r.status === 'fulfilled').length;
                const failed = results.filter(r => r.status === 'rejected').length;
                
                if (failed > 0) {
                    const errors = results
                        .filter(r => r.status === 'rejected')
                        .map(r => r.reason.message)
                        .join('\n');
                    alert(`Deleted ${successful} file(s), but ${failed} failed:\n\n${errors}`);
                } else {
                    console.log(`Successfully deleted ${successful} replay file(s)`);
                }
                
                deleteSelectedBtn.textContent = originalText;
                deleteSelectedBtn.disabled = false;
                browseDirectory(currentPath);
                // Update history status indicators if available
                if (window.updateHistoryStatus) {
                    window.updateHistoryStatus();
                }
            })
            .catch(err => {
                deleteSelectedBtn.textContent = originalText;
                deleteSelectedBtn.disabled = false;
                alert('Error deleting replays: ' + err.message);
            });
    }

    refreshReplaysBtn.addEventListener('click', () => browseDirectory(currentPath));

    selectAllBtn.addEventListener('click', () => {
        document.querySelectorAll('.replay-item input[type="checkbox"]').forEach(cb => {
            if (cb.value.endsWith('.dem')) cb.checked = true;
        });
        updatePlayerSelection();
    });

    deselectAllBtn.addEventListener('click', () => {
        document.querySelectorAll('.replay-item input[type="checkbox"]').forEach(cb => cb.checked = false);
        updatePlayerSelection();
    });

    selectLastBtn.addEventListener('click', () => {
        const count = parseInt(selectLastCountInput.value) || 10;
        const checkboxes = Array.from(document.querySelectorAll('.replay-item input[type="checkbox"]')).filter(cb => cb.value.endsWith('.dem'));
        checkboxes.forEach(cb => cb.checked = false);
        const newestX = checkboxes.slice(0, count);
        newestX.forEach(cb => cb.checked = true);
        updatePlayerSelection();
    });

    deleteSelectedBtn.addEventListener('click', deleteSelectedReplays);

    // Parse Logic
    startParseBtn.addEventListener('click', async () => {
        const selectedCheckboxes = document.querySelectorAll('.replay-item input[type="checkbox"]:checked');
        const selectedIds = Array.from(selectedCheckboxes).map(cb => cb.value);

        if (selectedIds.length === 0) {
            alert('Please select at least one replay.');
            return;
        }

        const steamIdToSend = "0";
        const slotId = -1;
        
        const steamIdValue = steamIdInput.value.trim();
        analysisSteamID = steamIdValue ? convertSteamIDTo64(steamIdValue) : null;
        console.log('Steam ID input value:', steamIdValue);
        console.log('Stored analysisSteamID (converted):', analysisSteamID);
        
        saveToStorage();

        // Reset UI
        progressSection.classList.remove('hidden');
        resultsSection.classList.add('hidden');
        graphsSection.classList.add('hidden');
        matchSelectorContainer.classList.add('hidden');
        playerSelectorContainer.classList.add('hidden');
        playerSelectorDisplay.innerHTML = '<span class="select-placeholder">Select a player...</span><span class="select-chevron">▼</span>';
        playerSelectorDropdown.classList.add('hidden');
        currentSelectedPlayer = null;
        playerSelectorOptions = [];
        progressBar.style.width = '0%';

        let totalTeamReports = 0;
        let totalEnemyReports = 0;
        let totalConfirmedTeamReports = 0;
        let totalConfirmedEnemyReports = 0;
        let processedCount = 0;
        
        const matchData = [];
        const confirmedPlayerReportCounts = new Map();
        const unconfirmedPlayerReportCounts = new Map();
        const confirmedTimelineData = [];
        const unconfirmedTimelineData = [];

        for (const filePath of selectedIds) {
            // Extract match ID from file path for display (e.g., "fatal/2025-11-17/8561630135.dem" -> "8561630135")
            const matchId = filePath.split('/').pop().replace('.dem', '');
            progressText.textContent = `Processing ${processedCount + 1} / ${selectedIds.length}: ${matchId}`;

            try {
                const res = await fetch('/api/parse', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        matchId: matchId,
                        filePath: filePath,
                        reportedSlot: -1,
                        reportedSteamId: "0",
                        profileName: getSelectedProfileName()
                    })
                });

                if (!res.ok) throw new Error(await res.text());

                const resultText = await res.text();
                const resultTextFixed = resultText.replace(/"TargetSteamID":\s*(\d+)/g, '"TargetSteamID":"$1"')
                    .replace(/"SteamID":\s*(\d+)/g, '"SteamID":"$1"');
                const result = JSON.parse(resultTextFixed);

                const countedSlots = new Set();
                let uniqueTeamReports = 0;
                let uniqueEnemyReports = 0;
                
                if (result.Reports) {
                    result.Reports.forEach(report => {
                        if (!countedSlots.has(report.Slot)) {
                            countedSlots.add(report.Slot);
                            if (report.Team === "FRIENDLY") {
                                uniqueTeamReports++;
                            } else {
                                uniqueEnemyReports++;
                            }
                        }
                    });
                }

                totalTeamReports += uniqueTeamReports;
                totalEnemyReports += uniqueEnemyReports;
                totalConfirmedTeamReports += uniqueTeamReports;
                totalConfirmedEnemyReports += uniqueEnemyReports;

                const confirmedTeamReports = uniqueTeamReports;
                const confirmedEnemyReports = uniqueEnemyReports;
                const unconfirmedTeamReports = 0;
                const unconfirmedEnemyReports = 0;

                matchData.push({
                    matchID: result.MatchID,
                    filePath: filePath, // Store filePath for later use (e.g., player-info)
                    teamReports: uniqueTeamReports,
                    enemyReports: uniqueEnemyReports,
                    confirmedTeamReports: confirmedTeamReports,
                    confirmedEnemyReports: confirmedEnemyReports,
                    unconfirmedTeamReports: unconfirmedTeamReports,
                    unconfirmedEnemyReports: unconfirmedEnemyReports,
                    reports: result.Reports || []
                });

                if (result.Reports) {
                    result.Reports.forEach(report => {
                        const playerKey = report.Name || `Slot ${report.Slot}`;
                        const timeParts = report.Time.split(':');
                        let totalMinutes = 0;
                        if (timeParts.length === 2) {
                            const minutes = parseInt(timeParts[0]) || 0;
                            const seconds = parseInt(timeParts[1]) || 0;
                            totalMinutes = minutes + seconds / 60;
                        }

                        const timelinePoint = {
                            x: totalMinutes,
                            y: report.Team === 'FRIENDLY' ? 1 : 2,
                            matchID: result.MatchID
                        };

                        if (report.Confirmed) {
                            confirmedPlayerReportCounts.set(playerKey, (confirmedPlayerReportCounts.get(playerKey) || 0) + 1);
                            confirmedTimelineData.push(timelinePoint);
                        } else {
                            unconfirmedPlayerReportCounts.set(playerKey, (unconfirmedPlayerReportCounts.get(playerKey) || 0) + 1);
                            unconfirmedTimelineData.push(timelinePoint);
                        }
                    });
                }

            } catch (err) {
                console.error(`Error processing ${matchId} (${filePath}):`, err);
            }

            processedCount++;
            progressBar.style.width = `${(processedCount / selectedIds.length) * 100}%`;
        }

        resultsSection.classList.remove('hidden');
        
        if (matchData.length > 0) {
            allMatchDataOriginal = JSON.parse(JSON.stringify(matchData));
            populateMatchSelector(matchData);
            if (matchData.length > 1) {
                matchSelectorContainer.classList.remove('hidden');
            } else {
                matchSelectorContainer.classList.add('hidden');
            }
        }

        if (totalTeamReports + totalEnemyReports > 0) {
            const totalUnconfirmedTeamReports = totalTeamReports - totalConfirmedTeamReports;
            const totalUnconfirmedEnemyReports = totalEnemyReports - totalConfirmedEnemyReports;
            
            allConfirmedPlayerReportCounts = confirmedPlayerReportCounts;
            allUnconfirmedPlayerReportCounts = unconfirmedPlayerReportCounts;
            allConfirmedTimelineData = confirmedTimelineData;
            allUnconfirmedTimelineData = unconfirmedTimelineData;
            allTotalTeamReports = totalTeamReports;
            allTotalEnemyReports = totalEnemyReports;
            allTotalConfirmedTeamReports = totalConfirmedTeamReports;
            allTotalConfirmedEnemyReports = totalConfirmedEnemyReports;
            allTotalUnconfirmedTeamReports = totalUnconfirmedTeamReports;
            allTotalUnconfirmedEnemyReports = totalUnconfirmedEnemyReports;
            
            updateGraphsForPlayer(currentSelectedPlayer);
            graphsSection.classList.remove('hidden');
        }
    });

    let timelineChartInstance = null;
    let allMatchData = [];
    let allMatchDataOriginal = [];
    let analysisSteamID = null;
    let playerReportsPerMatchChart = null;
    let currentSelectedPlayer = null;
    let allConfirmedPlayerReportCounts = new Map();
    let allUnconfirmedPlayerReportCounts = new Map();
    let allConfirmedTimelineData = [];
    let allUnconfirmedTimelineData = [];
    let allTotalTeamReports = 0;
    let allTotalEnemyReports = 0;
    let allTotalConfirmedTeamReports = 0;
    let allTotalConfirmedEnemyReports = 0;
    let allTotalUnconfirmedTeamReports = 0;
    let allTotalUnconfirmedEnemyReports = 0;

    function populateMatchSelector(matchData) {
        allMatchData = matchData;
        matchSelector.innerHTML = '<option value="">Select a match...</option>';
        matchData.forEach((match, index) => {
            const option = document.createElement('option');
            option.value = index;
            option.textContent = `Match ${match.matchID} (${match.teamReports + match.enemyReports} reports)`;
            matchSelector.appendChild(option);
        });
        if (matchData.length > 0) {
            matchSelector.value = '0';
            populatePlayerSelector(matchData[0]).then(() => {
                let steamIDToMatch = null;
                
                if (analysisSteamID) {
                    steamIDToMatch = String(analysisSteamID);
                } else {
                    const selectedProfile = getSelectedProfile();
                    if (selectedProfile && selectedProfile.id) {
                        steamIDToMatch = convertSteamIDTo64(String(selectedProfile.id));
                    }
                }
                
                if (steamIDToMatch) {
                    console.log('Initial match - Looking for Steam ID:', steamIDToMatch);
                    console.log('Available players:', playerSelectorOptions.map(opt => ({ key: opt.key, steamID: String(opt.steamID), name: opt.name })));
                    const matchingPlayer = playerSelectorOptions.find(opt => 
                        opt.steamID && String(opt.steamID) === steamIDToMatch
                    );
                    if (matchingPlayer) {
                        console.log('Found matching player, selecting:', matchingPlayer.key);
                        selectPlayerOption(matchingPlayer.key);
                    } else {
                        console.log('No matching player found for Steam ID:', steamIDToMatch);
                        renderTimelineGraph(matchData[0], null);
                    }
                } else {
                    renderTimelineGraph(matchData[0], null);
                }
            });
        }
    }

    async function populatePlayerSelector(matchData) {
        if (!matchData || !matchData.reports || matchData.reports.length === 0) {
            playerSelectorContainer.classList.add('hidden');
            return;
        }

        console.log('=== populatePlayerSelector DEBUG ===');
        console.log('Total reports:', matchData.reports.length);
        
        const uniqueTargetSlots = new Set();
        const uniqueTargetSteamIDs = new Set();
        const reportsByTargetSlot = new Map();
        const reportsByTargetSteamID = new Map();
        
        matchData.reports.forEach((r, idx) => {
            const targetSteamIDStr = r.TargetSteamID ? String(r.TargetSteamID) : null;
            if (r.TargetSlot != null) {
                uniqueTargetSlots.add(r.TargetSlot);
                if (!reportsByTargetSlot.has(r.TargetSlot)) {
                    reportsByTargetSlot.set(r.TargetSlot, []);
                }
                reportsByTargetSlot.get(r.TargetSlot).push({ idx, time: r.Time, targetSlot: r.TargetSlot, targetSteamID: targetSteamIDStr });
            }
            if (targetSteamIDStr) {
                uniqueTargetSteamIDs.add(targetSteamIDStr);
                if (!reportsByTargetSteamID.has(targetSteamIDStr)) {
                    reportsByTargetSteamID.set(targetSteamIDStr, []);
                }
                reportsByTargetSteamID.get(targetSteamIDStr).push({ idx, time: r.Time, targetSlot: r.TargetSlot, targetSteamID: targetSteamIDStr });
            }
        });
        console.log('Unique TargetSlots:', Array.from(uniqueTargetSlots).sort((a, b) => a - b));
        console.log('Unique TargetSteamIDs:', Array.from(uniqueTargetSteamIDs).sort((a, b) => a - b));
        
        const slot3SteamID = 76561198067621800;
        if (reportsByTargetSlot.has(3)) {
            console.log('Reports targeting slot 3:', reportsByTargetSlot.get(3));
        } else {
            console.log('No reports found with TargetSlot = 3');
        }
        if (reportsByTargetSteamID.has(slot3SteamID)) {
            console.log(`Reports targeting SteamID ${slot3SteamID} (slot 3):`, reportsByTargetSteamID.get(slot3SteamID));
        } else {
            console.log(`No reports found with TargetSteamID = ${slot3SteamID} (slot 3)`);
        }

        const buildPlayersMap = (steamIDToSlotMap) => {
            const map = new Map();
            const skipped = [];
            
            matchData.reports.forEach((report, idx) => {
                let key;
                const targetSlot = report.TargetSlot;
                const targetSteamID = report.TargetSteamID ? String(report.TargetSteamID) : null;
                
                let finalSlot = targetSlot;
                let finalSteamID = targetSteamID;
                
                if (targetSlot != null && targetSlot !== undefined && Number.isInteger(targetSlot) && targetSlot >= 0 && targetSlot < 10) {
                    key = `slot_${targetSlot}`;
                    if (!finalSteamID && slotToSteamID.has(targetSlot)) {
                        finalSteamID = slotToSteamID.get(targetSlot);
                    }
                } else if (targetSteamID && targetSteamID !== '0') {
                    if (steamIDToSlotMap && steamIDToSlotMap.has(targetSteamID)) {
                        finalSlot = steamIDToSlotMap.get(targetSteamID);
                        key = `slot_${finalSlot}`;
                    } else {
                        key = `steamid_${targetSteamID}`;
                    }
                } else {
                    skipped.push({ idx, targetSlot, targetSteamID, time: report.Time });
                    return;
                }
                
                if (!map.has(key)) {
                    map.set(key, {
                        slot: finalSlot,
                        steamID: finalSteamID,
                        reportCount: 0
                    });
                }
                map.get(key).reportCount++;
            });
            
            return { map, skipped };
        };
        
        const slotToSteamID = new Map();
        let steamIDToSlot = new Map();
        const skippedReports = [];
        
        matchData.reports.forEach((report, idx) => {
            const targetSlot = report.TargetSlot;
            const targetSteamID = report.TargetSteamID ? String(report.TargetSteamID) : null;
            
            if (targetSlot != null && targetSlot !== undefined && Number.isInteger(targetSlot) && targetSlot >= 0 && targetSlot < 10) {
                if (targetSteamID && targetSteamID !== '0') {
                    slotToSteamID.set(targetSlot, targetSteamID);
                    steamIDToSlot.set(targetSteamID, targetSlot);
                }
            }
        });
        
        console.log('Initial slotToSteamID mapping:', Array.from(slotToSteamID.entries()));
        console.log('Initial steamIDToSlot mapping:', Array.from(steamIDToSlot.entries()).map(([id, slot]) => `${id}: slot_${slot}`));
        
        let { map: playersMap, skipped: initialSkipped } = buildPlayersMap(steamIDToSlot);
        skippedReports.push(...initialSkipped);
        
        if (skippedReports.length > 0) {
            console.warn('Initial skipped reports:', skippedReports);
        }
        console.log('Initial playersMap created with', playersMap.size, 'players:', Array.from(playersMap.entries()).map(([k, v]) => `${k}: slot=${v.slot}, steamID=${v.steamID}, reports=${v.reportCount}`));

        try {
            const res = await fetch('/api/player-info', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    matchId: String(matchData.matchID),
                    filePath: matchData.filePath || '', // Include filePath if available
                    profileName: getSelectedProfileName()
                })
            });
            
            if (!res.ok) {
                throw new Error(`HTTP ${res.status}: ${await res.text()}`);
            }
            
            const playersText = await res.text();
            const playersTextFixed = playersText.replace(/"SteamID":\s*(\d+)/g, '"SteamID":"$1"');
            const players = JSON.parse(playersTextFixed);
            const playerInfoMap = new Map();
            players.forEach(p => {
                const steamIDStr = p.SteamID ? String(p.SteamID) : null;
                playerInfoMap.set(p.Slot, { name: p.Name, hero: p.Hero, team: p.Team, steamID: steamIDStr });
                if (steamIDStr && steamIDStr !== '0') {
                    steamIDToSlot.set(steamIDStr, p.Slot);
                    slotToSteamID.set(p.Slot, steamIDStr);
                }
            });
            
            console.log('Player info loaded:', Array.from(playerInfoMap.entries()).map(([slot, info]) => `slot_${slot}: ${info.name} (steamID: ${info.steamID})`));
            console.log('Updated steamIDToSlot mapping:', Array.from(steamIDToSlot.entries()).map(([id, slot]) => `${id}: slot_${slot}`));
            
            const { map: updatedPlayersMap, skipped: remainingSkipped } = buildPlayersMap(steamIDToSlot);
            playersMap = updatedPlayersMap;
            
            if (remainingSkipped.length > skippedReports.length) {
                const newlyMapped = remainingSkipped.length - skippedReports.length;
                console.log(`Rebuilt playersMap after player info. Previously skipped ${skippedReports.length}, now skipped ${remainingSkipped.length}`);
            }
            
            console.log('Final playersMap created with', playersMap.size, 'players:', Array.from(playersMap.entries()).map(([k, v]) => `${k}: slot=${v.slot}, steamID=${v.steamID}, reports=${v.reportCount}`));
            
            players.forEach(p => {
                if (p.Slot >= 0 && p.Slot < 10) {
                    const key = `slot_${p.Slot}`;
                    if (!playersMap.has(key)) {
                        console.log(`Adding missing slot ${p.Slot} (${p.Name}) with 0 reports`);
                        playersMap.set(key, {
                            slot: p.Slot,
                            steamID: p.SteamID,
                            reportCount: 0
                        });
                    }
                }
            });
            
            console.log('Final playersMap after adding missing slots:', playersMap.size, 'players:', Array.from(playersMap.entries()).map(([k, v]) => `${k}: slot=${v.slot}, steamID=${v.steamID}, reports=${v.reportCount}`));

            const targetPlayerMap = new Map();
            matchData.reports.forEach(report => {
                let key;
                const targetSlot = report.TargetSlot;
                const targetSteamID = report.TargetSteamID;
                
                let finalSlot = targetSlot;
                
                if (targetSlot != null && targetSlot !== undefined && Number.isInteger(targetSlot) && targetSlot >= 0 && targetSlot < 10) {
                    key = `slot_${targetSlot}`;
                } else if (targetSteamID && targetSteamID > 0) {
                    if (steamIDToSlot.has(targetSteamID)) {
                        finalSlot = steamIDToSlot.get(targetSteamID);
                        key = `slot_${finalSlot}`;
                    } else {
                        key = `steamid_${targetSteamID}`;
                    }
                } else {
                    return;
                }
                
                if (!targetPlayerMap.has(key)) {
                    targetPlayerMap.set(key, {
                        slot: finalSlot,
                        name: report.TargetName || '',
                        hero: report.TargetHero || ''
                    });
                }
            });
            
            console.log('targetPlayerMap keys after update:', Array.from(targetPlayerMap.keys()));

            playerSelectorOptions = [];
            const playerList = Array.from(playersMap.entries()).sort((a, b) => {
                const slotA = a[1].slot !== undefined && a[1].slot >= 0 ? a[1].slot : 999;
                const slotB = b[1].slot !== undefined && b[1].slot >= 0 ? b[1].slot : 999;
                return slotA - slotB;
            });
            
            playerSelectorOptions.push({ key: '', name: 'All Players', hero: '', team: null, reportCount: 0, slot: null });
            
            console.log('targetPlayerMap keys:', Array.from(targetPlayerMap.keys()));
            console.log('playerList keys:', playerList.map(([k]) => k));
            
            playerList.forEach(([key, player]) => {
                const targetInfo = targetPlayerMap.get(key);
                let apiInfo = null;
                if (player.slot !== undefined && player.slot >= 0 && player.slot < 10) {
                    apiInfo = playerInfoMap.get(player.slot);
                }
                const playerName = (targetInfo && targetInfo.name) || (apiInfo && apiInfo.name) || (player.slot !== undefined && player.slot >= 0 ? `Slot ${player.slot}` : 'Unknown Player');
                const heroName = (targetInfo && targetInfo.hero) || (apiInfo && apiInfo.hero) || '';
                const team = apiInfo && apiInfo.team ? (apiInfo.team === 2 ? 'radiant' : apiInfo.team === 3 ? 'dire' : null) : null;
                
                console.log(`Adding player option: key=${key}, slot=${player.slot}, steamID=${player.steamID}, name=${playerName}, hero=${heroName}, reports=${player.reportCount}`);
                
                playerSelectorOptions.push({
                    key: key,
                    name: playerName,
                    hero: heroName,
                    team: team,
                    reportCount: player.reportCount,
                    slot: player.slot,
                    steamID: player.steamID ? String(player.steamID) : null
                });
            });
            
            console.log('Final playerSelectorOptions count:', playerSelectorOptions.length);
            console.log('=== END populatePlayerSelector DEBUG ===');

            renderPlayerSelector();
            playerSelectorContainer.classList.remove('hidden');
        } catch (err) {
            console.error('Error loading player info:', err);
            const targetPlayerMap = new Map();
            matchData.reports.forEach(report => {
                let key;
                const targetSlot = report.TargetSlot;
                const targetSteamID = report.TargetSteamID;
                
                if (targetSlot != null && targetSlot !== undefined && Number.isInteger(targetSlot) && targetSlot >= 0 && targetSlot < 10) {
                    key = `slot_${targetSlot}`;
                } else if (targetSteamID && targetSteamID > 0) {
                    key = `steamid_${targetSteamID}`;
                } else {
                    return;
                }
                
                if (!targetPlayerMap.has(key)) {
                    targetPlayerMap.set(key, {
                        slot: targetSlot,
                        name: report.TargetName || '',
                        hero: report.TargetHero || ''
                    });
                }
            });

            playerSelectorOptions = [];
            const playerList = Array.from(playersMap.entries()).sort((a, b) => {
                const slotA = a[1].slot !== undefined && a[1].slot >= 0 && a[1].slot < 10 ? a[1].slot : 999;
                const slotB = b[1].slot !== undefined && b[1].slot >= 0 && b[1].slot < 10 ? b[1].slot : 999;
                return slotA - slotB;
            });
            
            playerSelectorOptions.push({ key: '', name: 'All Players', hero: '', team: null, reportCount: 0, slot: null });
            
            playerList.forEach(([key, player]) => {
                const targetInfo = targetPlayerMap.get(key);
                const playerName = (targetInfo && targetInfo.name) || (player.slot !== undefined && player.slot >= 0 ? `Slot ${player.slot}` : 'Unknown Player');
                const heroName = (targetInfo && targetInfo.hero) || '';
                
                playerSelectorOptions.push({
                    key: key,
                    name: playerName,
                    hero: heroName,
                    team: null,
                    reportCount: player.reportCount,
                    slot: player.slot,
                    steamID: player.steamID ? String(player.steamID) : null
                });
            });

            renderPlayerSelector();
            playerSelectorContainer.classList.remove('hidden');
        }
    }

    function parseTimeToMinutes(timeStr) {
        const parts = timeStr.split(':');
        if (parts.length !== 2) return 0;
        const minutes = parseInt(parts[0]) || 0;
        const seconds = parseInt(parts[1]) || 0;
        return minutes + seconds / 60;
    }

    function renderTimelineGraph(matchData, playerFilter = null) {
        if (!matchData || !matchData.reports || matchData.reports.length === 0) {
            timelineGraphContainer.innerHTML = '<p style="text-align: center; color: var(--text-secondary); padding: 2rem;">No reports found for this match.</p>';
            return;
        }

        let reports = matchData.reports;
        if (playerFilter) {
            if (playerFilter.startsWith('slot_')) {
                const targetSlot = parseInt(playerFilter.replace('slot_', ''));
                reports = reports.filter(r => r.TargetSlot === targetSlot);
            } else if (playerFilter.startsWith('steamid_')) {
                const targetSteamID = playerFilter.replace('steamid_', '');
                reports = reports.filter(r => String(r.TargetSteamID) === targetSteamID);
            }
        }

        if (reports.length === 0) {
            timelineGraphContainer.innerHTML = '<p style="text-align: center; color: var(--text-secondary); padding: 2rem;">No reports found for the selected player.</p>';
            return;
        }

        const canvas = document.createElement('canvas');
        canvas.id = 'timeline-graph';
        const tooltip = document.createElement('div');
        tooltip.id = 'timeline-tooltip';
        tooltip.className = 'timeline-tooltip hidden';
        timelineGraphContainer.innerHTML = '';
        timelineGraphContainer.appendChild(canvas);
        timelineGraphContainer.appendChild(tooltip);

        const ctx = canvas.getContext('2d');
        const padding = { top: 60, right: 40, bottom: 60, left: 80 };
        const iconSize = 40;
        const iconSpacing = 5;
        const teamRegionHeight = 120;
        const totalHeight = padding.top + teamRegionHeight * 2 + padding.bottom;
        
        const containerWidth = Math.max(timelineGraphContainer.clientWidth || 1200, 800);
        const devicePixelRatio = window.devicePixelRatio || 1;
        canvas.width = containerWidth * devicePixelRatio;
        canvas.height = totalHeight * devicePixelRatio;
        canvas.style.width = '100%';
        canvas.style.height = `${totalHeight}px`;
        ctx.scale(devicePixelRatio, devicePixelRatio);
        
        const graphWidth = containerWidth - padding.left - padding.right;
        const graphHeight = totalHeight - padding.top - padding.bottom;

        const maxTime = Math.max(...reports.map(r => parseTimeToMinutes(r.Time)), 60);
        const timeRange = Math.max(maxTime, 60);

        const friendlyReports = reports.filter(r => r.Team === 'FRIENDLY');
        const enemyReports = reports.filter(r => r.Team === 'ENEMY');

        function groupReportsByTime(reportList) {
            const grouped = new Map();
            reportList.forEach(report => {
                const time = parseTimeToMinutes(report.Time);
                const timeKey = Math.floor(time * 60);
                if (!grouped.has(timeKey)) {
                    grouped.set(timeKey, []);
                }
                grouped.get(timeKey).push({ ...report, time });
            });
            return grouped;
        }

        const friendlyGroups = groupReportsByTime(friendlyReports);
        const enemyGroups = groupReportsByTime(enemyReports);

        function getCircleCenter(pos) {
            return {
                x: pos.x,
                y: pos.y + iconSize / 2
            };
        }

        function getDistance(pos1, pos2) {
            const center1 = getCircleCenter(pos1);
            const center2 = getCircleCenter(pos2);
            const dx = center1.x - center2.x;
            const dy = center1.y - center2.y;
            return Math.sqrt(dx * dx + dy * dy);
        }

        function resolveOverlaps(positions, radius, minY, maxY) {
            const sorted = [...positions].sort((a, b) => {
                const xDiff = Math.abs(a.x - b.x);
                if (xDiff < 1) {
                    return a.y - b.y;
                }
                return a.x - b.x;
            });

            const xThreshold = radius * 1.5;
            const minSeparation = radius * 2 + iconSpacing;

            const groups = new Map();
            sorted.forEach((pos, idx) => {
                let foundGroup = false;
                for (const [groupX, group] of groups.entries()) {
                    if (Math.abs(pos.x - groupX) < xThreshold) {
                        group.push(idx);
                        foundGroup = true;
                        break;
                    }
                }
                if (!foundGroup) {
                    groups.set(pos.x, [idx]);
                }
            });

            groups.forEach((indices) => {
                indices.sort((a, b) => sorted[a].y - sorted[b].y);
                let currentY = minY + 10;
                indices.forEach(idx => {
                    sorted[idx].y = currentY;
                    currentY += iconSize + iconSpacing;
                });
            });

            const maxIterations = 500;
            const damping = 0.9;
            const minForce = 0.1;

            for (let iter = 0; iter < maxIterations; iter++) {
                let totalMovement = 0;
                const forces = new Array(sorted.length).fill(0);

                for (let i = 0; i < sorted.length; i++) {
                    for (let j = i + 1; j < sorted.length; j++) {
                        const distance = getDistance(sorted[i], sorted[j]);
                        
                        if (distance < minSeparation) {
                            const center1 = getCircleCenter(sorted[i]);
                            const center2 = getCircleCenter(sorted[j]);
                            const dx = center1.x - center2.x;
                            const dy = center1.y - center2.y;
                            
                            if (distance > 0.001) {
                                const overlap = minSeparation - distance;
                                const force = (overlap / distance) * 1.5;
                                const forceY = (dy / distance) * force;
                                
                                forces[i] += forceY;
                                forces[j] -= forceY;
                            } else {
                                const pushY = (minSeparation / 2) + iconSpacing;
                                if (sorted[i].y < sorted[j].y) {
                                    forces[i] -= pushY;
                                    forces[j] += pushY;
                                } else {
                                    forces[i] += pushY;
                                    forces[j] -= pushY;
                                }
                            }
                        }
                    }
                }

                for (let i = 0; i < sorted.length; i++) {
                    if (Math.abs(forces[i]) > minForce) {
                        const newY = sorted[i].y + forces[i] * damping;
                        const clampedY = Math.max(minY, Math.min(maxY - iconSize, newY));
                        const movement = Math.abs(clampedY - sorted[i].y);
                        totalMovement += movement;
                        sorted[i].y = clampedY;
                    }
                }

                if (totalMovement < 0.01) break;
            }

            for (let i = 0; i < sorted.length; i++) {
                for (let j = i + 1; j < sorted.length; j++) {
                    const distance = getDistance(sorted[i], sorted[j]);
                    if (distance < minSeparation - 0.5) {
                        const center1 = getCircleCenter(sorted[i]);
                        const center2 = getCircleCenter(sorted[j]);
                        const dy = center1.y - center2.y;
                        const neededSeparation = minSeparation - distance;
                        
                        if (Math.abs(dy) < iconSize) {
                            if (sorted[i].y < sorted[j].y) {
                                sorted[i].y = Math.max(minY, sorted[i].y - neededSeparation / 2);
                                sorted[j].y = Math.min(maxY - iconSize, sorted[j].y + neededSeparation / 2);
                            } else {
                                sorted[j].y = Math.max(minY, sorted[j].y - neededSeparation / 2);
                                sorted[i].y = Math.min(maxY - iconSize, sorted[i].y + neededSeparation / 2);
                            }
                        }
                    }
                }
            }

            return sorted;
        }

        const iconPositions = [];
        const friendlyYByX = new Map();
        const enemyYByX = new Map();
        
        friendlyGroups.forEach((reportGroup) => {
            const reports = reportGroup.sort((a, b) => a.time - b.time);
            const time = reports[0].time;
            const x = padding.left + (time / timeRange) * graphWidth;
            
            let currentY = friendlyYByX.get(x) || padding.top + 10;
            
            reports.forEach((report) => {
                iconPositions.push({ report, x, y: currentY, team: 'FRIENDLY' });
                currentY += iconSize + iconSpacing;
            });
            
            friendlyYByX.set(x, currentY);
        });
        
        enemyGroups.forEach((reportGroup) => {
            const reports = reportGroup.sort((a, b) => a.time - b.time);
            const time = reports[0].time;
            const x = padding.left + (time / timeRange) * graphWidth;
            
            let currentY = enemyYByX.get(x) || padding.top + teamRegionHeight + 10;
            
            reports.forEach((report) => {
                iconPositions.push({ report, x, y: currentY, team: 'ENEMY' });
                currentY += iconSize + iconSpacing;
            });
            
            enemyYByX.set(x, currentY);
        });

        const friendlyPositions = iconPositions.filter(p => p.team === 'FRIENDLY');
        const enemyPositions = iconPositions.filter(p => p.team === 'ENEMY');
        
        const radius = iconSize / 2;
        const friendlyResolved = resolveOverlaps(friendlyPositions, radius, padding.top, padding.top + teamRegionHeight);
        const enemyResolved = resolveOverlaps(enemyPositions, radius, padding.top + teamRegionHeight, padding.top + teamRegionHeight * 2);
        
        iconPositions.length = 0;
        iconPositions.push(...friendlyResolved, ...enemyResolved);

        const imagesToLoad = new Set();
        iconPositions.forEach(pos => {
            const iconUrl = getHeroIconUrl(pos.report.TargetHero || pos.report.Hero);
            if (iconUrl) imagesToLoad.add(iconUrl);
        });

        const imageCache = new Map();
        let imagesLoaded = 0;
        const totalImages = imagesToLoad.size;

        function drawIcons() {
            ctx.clearRect(0, 0, containerWidth, totalHeight);
            ctx.fillStyle = '#0f172a';
            ctx.fillRect(0, 0, containerWidth, totalHeight);

            ctx.strokeStyle = '#334155';
            ctx.lineWidth = 1;
            ctx.beginPath();
            ctx.moveTo(padding.left, padding.top);
            ctx.lineTo(padding.left, totalHeight - padding.bottom);
            ctx.moveTo(padding.left, padding.top + teamRegionHeight);
            ctx.lineTo(containerWidth - padding.right, padding.top + teamRegionHeight);
            ctx.moveTo(padding.left, padding.top + teamRegionHeight * 2);
            ctx.lineTo(containerWidth - padding.right, padding.top + teamRegionHeight * 2);
            ctx.stroke();

            ctx.fillStyle = '#94a3b8';
            ctx.font = '12px Inter';
            ctx.textAlign = 'center';
            ctx.fillText('FRIENDLY', containerWidth / 2, padding.top - 20);
            ctx.fillText('ENEMY', containerWidth / 2, padding.top + teamRegionHeight + 20);

            const numLabels = Math.min(20, Math.ceil(timeRange / 2));
            for (let i = 0; i <= numLabels; i++) {
                const time = (i * timeRange) / numLabels;
                const x = padding.left + (time / timeRange) * graphWidth;
                const minutes = Math.floor(time);
                const seconds = Math.floor((time - minutes) * 60);
                const timeLabel = `${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`;
                
                ctx.strokeStyle = '#475569';
                ctx.lineWidth = 0.5;
                ctx.beginPath();
                ctx.moveTo(x, padding.top);
                ctx.lineTo(x, totalHeight - padding.bottom);
                ctx.stroke();
                
                ctx.fillStyle = '#94a3b8';
                ctx.font = '11px Inter';
                ctx.textAlign = 'center';
                ctx.fillText(timeLabel, x, totalHeight - padding.bottom + 18);
            }

            iconPositions.forEach(({ report, x, y, team }) => {
                const heroName = report.TargetHero || report.Hero;
                const iconUrl = getHeroIconUrl(heroName);
                const img = iconUrl ? imageCache.get(iconUrl) : null;
                
                if (img && img.complete && img.naturalWidth > 0 && img.naturalHeight > 0) {
                    ctx.save();
                    ctx.beginPath();
                    ctx.arc(x, y + iconSize / 2, iconSize / 2, 0, Math.PI * 2);
                    ctx.clip();
                    ctx.drawImage(img, x - iconSize / 2, y, iconSize, iconSize);
                    ctx.restore();
                } else {
                    ctx.fillStyle = '#475569';
                    ctx.beginPath();
                    ctx.arc(x, y + iconSize / 2, iconSize / 2, 0, Math.PI * 2);
                    ctx.fill();
                    ctx.fillStyle = '#f1f5f9';
                    ctx.font = '10px Inter';
                    ctx.textAlign = 'center';
                    ctx.fillText(heroName ? heroName.substring(0, 3).toUpperCase() : '?', x, y + iconSize / 2 + 3);
                }
                
                ctx.strokeStyle = team === 'FRIENDLY' ? '#22c55e' : '#ef4444';
                ctx.lineWidth = 2;
                ctx.beginPath();
                ctx.arc(x, y + iconSize / 2, iconSize / 2, 0, Math.PI * 2);
                ctx.stroke();
            });
        }

        drawIcons();

        if (totalImages > 0) {
            imagesToLoad.forEach(iconUrl => {
                const img = new Image();
                img.onload = () => {
                    imageCache.set(iconUrl, img);
                    imagesLoaded++;
                    if (imagesLoaded === totalImages) {
                        drawIcons();
                    } else if (imagesLoaded % 3 === 0) {
                        drawIcons();
                    }
                };
                img.onerror = (e) => {
                    console.warn(`Failed to load hero icon: ${iconUrl}`, e);
                    imagesLoaded++;
                    if (imagesLoaded === totalImages) {
                        drawIcons();
                    }
                };
                img.src = iconUrl;
            });
        }

        canvas.addEventListener('mousemove', (e) => {
            const rect = canvas.getBoundingClientRect();
            const scaleX = rect.width / containerWidth;
            const scaleY = rect.height / totalHeight;
            const x = (e.clientX - rect.left) / scaleX;
            const y = (e.clientY - rect.top) / scaleY;
            
            const iconRadius = iconSize / 2;
            let hoveredIcon = null;
            
            for (const pos of iconPositions) {
                const centerX = pos.x;
                const centerY = pos.y + iconSize / 2;
                const dx = x - centerX;
                const dy = y - centerY;
                const distance = Math.sqrt(dx * dx + dy * dy);
                
                if (distance <= iconRadius) {
                    hoveredIcon = pos;
                    break;
                }
            }
            
            if (hoveredIcon) {
                const timestamp = hoveredIcon.report.Time;
                const heroName = hoveredIcon.report.TargetHero || hoveredIcon.report.Hero;
                tooltip.textContent = `${heroName || 'Unknown'}: ${timestamp}`;
                tooltip.style.visibility = 'hidden';
                tooltip.classList.remove('hidden');
                
                const tooltipRect = tooltip.getBoundingClientRect();
                let left = e.clientX + 15;
                let top = e.clientY - tooltipRect.height - 10;
                
                if (left + tooltipRect.width > window.innerWidth - 10) {
                    left = e.clientX - tooltipRect.width - 15;
                }
                if (top < 10) {
                    top = e.clientY + 15;
                }
                
                tooltip.style.left = `${left}px`;
                tooltip.style.top = `${top}px`;
                tooltip.style.visibility = 'visible';
            } else {
                tooltip.classList.add('hidden');
            }
        });
        
        canvas.addEventListener('mouseleave', () => {
            tooltip.classList.add('hidden');
        });
    }

    function renderPlayerSelector() {
        playerSelectorDropdown.innerHTML = '';
        
        playerSelectorOptions.forEach(option => {
            const optionEl = document.createElement('div');
            optionEl.className = 'custom-select-option';
            optionEl.dataset.key = option.key;
            
            const iconUrl = option.hero ? getHeroIconUrl(option.hero) : null;
            const playerColor = getSlotColor(option.slot);
            
            let html = '';
            if (iconUrl) {
                html += `<img src="${iconUrl}" alt="${option.hero}" class="player-selector-icon" onerror="this.style.display='none'">`;
            }
            html += `<span class="player-selector-name" style="color: ${playerColor}">${option.name}</span>`;
            if (option.reportCount > 0) {
                html += `<span class="player-selector-count">${option.reportCount} report${option.reportCount !== 1 ? 's' : ''}</span>`;
            }
            
            optionEl.innerHTML = html;
            optionEl.addEventListener('click', () => {
                selectPlayerOption(option.key);
                playerSelectorDropdown.classList.add('hidden');
            });
            playerSelectorDropdown.appendChild(optionEl);
        });
    }

    function selectPlayerOption(key) {
        currentSelectedPlayer = key || null;
        const option = playerSelectorOptions.find(o => o.key === key);
        
        if (option) {
            const iconUrl = option.hero ? getHeroIconUrl(option.hero) : null;
            const playerColor = getSlotColor(option.slot);
            
            let html = '';
            if (iconUrl) {
                html += `<img src="${iconUrl}" alt="${option.hero}" class="player-selector-icon" onerror="this.style.display='none'">`;
            }
            html += `<span class="player-selector-name" style="color: ${playerColor}">${option.name}</span>`;
            if (option.reportCount > 0) {
                html += `<span class="player-selector-count">${option.reportCount} report${option.reportCount !== 1 ? 's' : ''}</span>`;
            }
            html += '<span class="select-chevron">▼</span>';
            playerSelectorDisplay.innerHTML = html;
        } else {
            playerSelectorDisplay.innerHTML = '<span class="select-placeholder">Select a player...</span><span class="select-chevron">▼</span>';
        }
        
        const matchIndex = parseInt(matchSelector.value);
        if (matchIndex >= 0 && matchIndex < allMatchData.length) {
            renderTimelineGraph(allMatchData[matchIndex], currentSelectedPlayer);
            updateGraphsForPlayer(currentSelectedPlayer);
        }
    }

    playerSelectorDisplay.addEventListener('click', (e) => {
        e.stopPropagation();
        playerSelectorDropdown.classList.toggle('hidden');
    });

    document.addEventListener('click', (e) => {
        if (!playerSelectorContainer.contains(e.target)) {
            playerSelectorDropdown.classList.add('hidden');
        }
    });

    matchSelector.addEventListener('change', (e) => {
        const index = parseInt(e.target.value);
        if (index >= 0 && index < allMatchData.length) {
            populatePlayerSelector(allMatchData[index]).then(() => {
                let steamIDToMatch = null;
                
                if (analysisSteamID) {
                    steamIDToMatch = String(analysisSteamID);
                    console.log('Using analysisSteamID:', steamIDToMatch);
                } else {
                    const selectedProfile = getSelectedProfile();
                    if (selectedProfile && selectedProfile.id) {
                        steamIDToMatch = convertSteamIDTo64(String(selectedProfile.id));
                        console.log('Using profile SteamID (converted):', steamIDToMatch);
                    }
                }
                
                if (steamIDToMatch) {
                    console.log('Looking for Steam ID:', steamIDToMatch);
                    console.log('Available players:', playerSelectorOptions.map(opt => ({ key: opt.key, steamID: String(opt.steamID), name: opt.name })));
                    const matchingPlayer = playerSelectorOptions.find(opt => 
                        opt.steamID && String(opt.steamID) === steamIDToMatch
                    );
                    if (matchingPlayer) {
                        console.log('Found matching player, selecting:', matchingPlayer.key);
                        selectPlayerOption(matchingPlayer.key);
                    } else {
                        console.log('No matching player found for Steam ID:', steamIDToMatch);
                        const slot0Player = playerSelectorOptions.find(opt => opt.slot === 0);
                        if (slot0Player) {
                            console.log('Falling back to slot 0:', slot0Player);
                            selectPlayerOption(slot0Player.key);
                        } else {
                            renderTimelineGraph(allMatchData[index], null);
                            updateGraphsForPlayer(null);
                        }
                    }
                } else {
                    console.log('No Steam ID to match');
                    renderTimelineGraph(allMatchData[index], null);
                    updateGraphsForPlayer(null);
                }
            });
        }
    });


    let chartInstances = [];

    function filterDataByPlayer(matchData, playerFilter) {
        if (!playerFilter) return matchData;
        
        const filteredMatchData = matchData.map(match => {
            let filteredReports = match.reports || [];
            if (playerFilter.startsWith('slot_')) {
                const targetSlot = parseInt(playerFilter.replace('slot_', ''));
                filteredReports = filteredReports.filter(r => r.TargetSlot === targetSlot);
            } else if (playerFilter.startsWith('steamid_')) {
                const targetSteamID = playerFilter.replace('steamid_', '');
                filteredReports = filteredReports.filter(r => String(r.TargetSteamID) === targetSteamID);
            }
            
            const countedSlots = new Set();
            let teamReports = 0;
            let enemyReports = 0;
            filteredReports.forEach(report => {
                if (!countedSlots.has(report.Slot)) {
                    countedSlots.add(report.Slot);
                    if (report.Team === "FRIENDLY") {
                        teamReports++;
                    } else {
                        enemyReports++;
                    }
                }
            });
            
            return {
                ...match,
                reports: filteredReports,
                teamReports: teamReports,
                enemyReports: enemyReports,
                confirmedTeamReports: teamReports,
                confirmedEnemyReports: enemyReports,
                unconfirmedTeamReports: 0,
                unconfirmedEnemyReports: 0
            };
        });
        
        return filteredMatchData;
    }

    function filterReportCountsByPlayer(reportCounts, matchData, playerFilter) {
        if (!playerFilter) return reportCounts;
        
        const filtered = new Map();
        const filteredMatchData = filterDataByPlayer(matchData, playerFilter);
        
        filteredMatchData.forEach(match => {
            (match.reports || []).forEach(report => {
                const playerKey = report.Name || `Slot ${report.Slot}`;
                filtered.set(playerKey, (filtered.get(playerKey) || 0) + 1);
            });
        });
        
        return filtered;
    }

    function filterTimelineDataByPlayer(timelineData, matchData, playerFilter) {
        if (!playerFilter) return timelineData;
        
        const filteredMatchData = filterDataByPlayer(matchData, playerFilter);
        const filtered = [];
        
        filteredMatchData.forEach(match => {
            (match.reports || []).forEach(report => {
                const timeParts = report.Time.split(':');
                let totalMinutes = 0;
                if (timeParts.length === 2) {
                    const minutes = parseInt(timeParts[0]) || 0;
                    const seconds = parseInt(timeParts[1]) || 0;
                    totalMinutes = minutes + seconds / 60;
                }
                filtered.push({
                    x: totalMinutes,
                    y: report.Team === 'FRIENDLY' ? 1 : 2,
                    matchID: match.matchID
                });
            });
        });
        
        return filtered;
    }

    function calculateTotalsForPlayer(matchData, playerFilter) {
        if (!playerFilter) {
            return {
                totalTeamReports: allTotalTeamReports,
                totalEnemyReports: allTotalEnemyReports,
                totalConfirmedTeamReports: allTotalConfirmedTeamReports,
                totalConfirmedEnemyReports: allTotalConfirmedEnemyReports,
                totalUnconfirmedTeamReports: allTotalUnconfirmedTeamReports,
                totalUnconfirmedEnemyReports: allTotalUnconfirmedEnemyReports
            };
        }
        
        const filteredMatchData = filterDataByPlayer(matchData, playerFilter);
        let totalTeamReports = 0;
        let totalEnemyReports = 0;
        
        filteredMatchData.forEach(match => {
            totalTeamReports += match.teamReports || 0;
            totalEnemyReports += match.enemyReports || 0;
        });
        
        return {
            totalTeamReports: totalTeamReports,
            totalEnemyReports: totalEnemyReports,
            totalConfirmedTeamReports: totalTeamReports,
            totalConfirmedEnemyReports: totalEnemyReports,
            totalUnconfirmedTeamReports: 0,
            totalUnconfirmedEnemyReports: 0
        };
    }

    function updateGraphsForPlayer(playerFilter) {
        const matchIndex = parseInt(matchSelector.value);
        if (matchIndex >= 0 && matchIndex < allMatchData.length) {
            const selectedMatch = allMatchData[matchIndex];
            const filteredMatchData = filterDataByPlayer([selectedMatch], playerFilter);
            const filteredConfirmedCounts = filterReportCountsByPlayer(allConfirmedPlayerReportCounts, [selectedMatch], playerFilter);
            const filteredUnconfirmedCounts = filterReportCountsByPlayer(allUnconfirmedPlayerReportCounts, [selectedMatch], playerFilter);
            const filteredConfirmedTimeline = filterTimelineDataByPlayer(allConfirmedTimelineData, [selectedMatch], playerFilter);
            const filteredUnconfirmedTimeline = filterTimelineDataByPlayer(allUnconfirmedTimelineData, [selectedMatch], playerFilter);
            const totals = calculateTotalsForPlayer([selectedMatch], playerFilter);
            
            generateGraphs(filteredMatchData, filteredConfirmedCounts, filteredUnconfirmedCounts,
                filteredConfirmedTimeline, filteredUnconfirmedTimeline,
                totals.totalTeamReports, totals.totalEnemyReports,
                totals.totalConfirmedTeamReports, totals.totalConfirmedEnemyReports,
                totals.totalUnconfirmedTeamReports, totals.totalUnconfirmedEnemyReports);
        } else {
            const filteredMatchData = filterDataByPlayer(allMatchData, playerFilter);
            const filteredConfirmedCounts = filterReportCountsByPlayer(allConfirmedPlayerReportCounts, allMatchData, playerFilter);
            const filteredUnconfirmedCounts = filterReportCountsByPlayer(allUnconfirmedPlayerReportCounts, allMatchData, playerFilter);
            const filteredConfirmedTimeline = filterTimelineDataByPlayer(allConfirmedTimelineData, allMatchData, playerFilter);
            const filteredUnconfirmedTimeline = filterTimelineDataByPlayer(allUnconfirmedTimelineData, allMatchData, playerFilter);
            const totals = calculateTotalsForPlayer(allMatchData, playerFilter);
            
            generateGraphs(filteredMatchData, filteredConfirmedCounts, filteredUnconfirmedCounts,
                filteredConfirmedTimeline, filteredUnconfirmedTimeline,
                totals.totalTeamReports, totals.totalEnemyReports,
                totals.totalConfirmedTeamReports, totals.totalConfirmedEnemyReports,
                totals.totalUnconfirmedTeamReports, totals.totalUnconfirmedEnemyReports);
        }
        
        if (allMatchDataOriginal.length > 1 && analysisSteamID) {
            updatePlayerReportsPerMatchGraph();
        } else {
            const playerReportsSection = document.getElementById('playerReportsPerMatchSection');
            if (playerReportsSection) {
                playerReportsSection.style.display = 'none';
            }
        }
    }
    
    function updatePlayerReportsPerMatchGraph() {
        const playerReportsSection = document.getElementById('playerReportsPerMatchSection');
        if (!playerReportsSection) return;
        
        let targetSteamID = null;
        
        if (analysisSteamID) {
            targetSteamID = String(analysisSteamID);
        } else {
            const selectedProfile = getSelectedProfile();
            if (selectedProfile && selectedProfile.id) {
                targetSteamID = convertSteamIDTo64(String(selectedProfile.id));
            }
        }
        
        if (!targetSteamID || targetSteamID === '0' || targetSteamID === 'null' || targetSteamID === 'undefined') {
            playerReportsSection.style.display = 'none';
            return;
        }
        
        console.log('Player Reports Per Match - Using Steam ID:', targetSteamID);
        
        playerReportsSection.style.display = 'block';
        
        const reportsPerMatch = allMatchDataOriginal.map(match => {
            const reports = match.reports.filter(r => {
                const reportSteamID = r.TargetSteamID ? String(r.TargetSteamID) : null;
                return reportSteamID && reportSteamID === targetSteamID;
            });
            const uniqueReporters = new Set();
            reports.forEach(report => {
                if (report.Slot != null) {
                    uniqueReporters.add(report.Slot);
                }
            });
            return uniqueReporters.size;
        });
        
        const matchLabels = allMatchDataOriginal.map(m => `Match ${m.matchID}`);
        
        if (playerReportsPerMatchChart) {
            playerReportsPerMatchChart.destroy();
        }
        
        const ctx = document.getElementById('playerReportsPerMatchChart').getContext('2d');
        playerReportsPerMatchChart = new Chart(ctx, {
            type: 'bar',
            data: {
                labels: matchLabels,
                datasets: [{
                    label: 'Unique Reporters',
                    data: reportsPerMatch,
                    backgroundColor: '#3b82f6',
                    borderColor: '#2563eb',
                    borderWidth: 1
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: {
                        labels: {
                            color: '#f1f5f9'
                        }
                    },
                    title: {
                        display: true,
                        text: `Unique Reporters Per Match (Steam ID: ${targetSteamID})`,
                        color: '#f1f5f9',
                        font: { size: 14, weight: 'bold' }
                    }
                },
                scales: {
                    x: {
                        ticks: { color: '#94a3b8' },
                        grid: { color: '#334155' }
                    },
                    y: {
                        ticks: { 
                            color: '#94a3b8',
                            stepSize: 1,
                            precision: 0
                        },
                        grid: { color: '#334155' },
                        beginAtZero: true
                    }
                }
            }
        });
    }

    function generateGraphs(matchData, confirmedPlayerReportCounts, unconfirmedPlayerReportCounts, 
        confirmedTimelineData, unconfirmedTimelineData,
        totalTeamReports, totalEnemyReports, 
        totalConfirmedTeamReports, totalConfirmedEnemyReports,
        totalUnconfirmedTeamReports, totalUnconfirmedEnemyReports) {
        chartInstances.forEach(chart => chart.destroy());
        chartInstances = [];

        // Calculate and display total unique report counts for current match/selection
        const totalFriendly = totalTeamReports;
        const totalEnemy = totalEnemyReports;
        const totalCombined = totalFriendly + totalEnemy;
        
        const reportSummary = document.getElementById('report-summary');
        if (reportSummary && totalCombined > 0) {
            document.getElementById('total-friendly-reports').textContent = totalFriendly;
            document.getElementById('total-enemy-reports').textContent = totalEnemy;
            document.getElementById('total-combined-reports').textContent = totalCombined;
            reportSummary.style.display = 'block';
        } else if (reportSummary) {
            reportSummary.style.display = 'none';
        }

        // Calculate and display totals for all matches by summing unique reports received by the target Steam ID
        let allMatchesFriendly = 0;
        let allMatchesEnemy = 0;
        
        if (allMatchDataOriginal && allMatchDataOriginal.length > 0 && analysisSteamID) {
            allMatchDataOriginal.forEach(match => {
                // Filter reports to only count those targeting the analysis Steam ID
                const targetReports = (match.reports || []).filter(r => {
                    const targetSteamID = r.TargetSteamID ? String(r.TargetSteamID) : null;
                    return targetSteamID === analysisSteamID;
                });
                
                // Count unique reports by Slot (reporter slot) for the target player
                const countedSlots = new Set();
                let friendly = 0;
                let enemy = 0;
                
                targetReports.forEach(report => {
                    if (!countedSlots.has(report.Slot)) {
                        countedSlots.add(report.Slot);
                        if (report.Team === "FRIENDLY") {
                            friendly++;
                        } else {
                            enemy++;
                        }
                    }
                });
                
                allMatchesFriendly += friendly;
                allMatchesEnemy += enemy;
            });
        }
        
        const allMatchesCombined = allMatchesFriendly + allMatchesEnemy;
        
        const allMatchesReportSummary = document.getElementById('all-matches-report-summary');
        if (allMatchesReportSummary && allMatchesCombined > 0) {
            document.getElementById('all-matches-friendly-reports').textContent = allMatchesFriendly;
            document.getElementById('all-matches-enemy-reports').textContent = allMatchesEnemy;
            document.getElementById('all-matches-combined-reports').textContent = allMatchesCombined;
            allMatchesReportSummary.style.display = 'block';
        } else if (allMatchesReportSummary) {
            allMatchesReportSummary.style.display = 'none';
        }

        const chartOptions = {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: {
                    labels: {
                        color: '#f1f5f9'
                    }
                }
            },
            scales: {
                x: {
                    ticks: { color: '#94a3b8' },
                    grid: { color: '#334155' }
                },
                y: {
                    ticks: { color: '#94a3b8' },
                    grid: { color: '#334155' }
                }
            }
        };

        const hasConfirmed = totalConfirmedTeamReports + totalConfirmedEnemyReports > 0;
        const hasUnconfirmed = totalUnconfirmedTeamReports + totalUnconfirmedEnemyReports > 0;

        document.querySelectorAll('[id^="confirmed"]').forEach(el => {
            if (el.tagName === 'CANVAS') {
                el.parentElement.style.display = hasConfirmed ? 'block' : 'none';
            }
        });
        document.querySelectorAll('[id^="unconfirmed"]').forEach(el => {
            if (el.tagName === 'CANVAS') {
                el.parentElement.style.display = hasUnconfirmed ? 'block' : 'none';
            }
        });
        
        const unconfirmedTitle = document.querySelector('.unconfirmed-section-title');
        if (unconfirmedTitle) {
            unconfirmedTitle.style.display = hasUnconfirmed ? 'block' : 'none';
        }

        if (totalConfirmedTeamReports + totalConfirmedEnemyReports > 0) {
            const confirmedTeamEnemyCtx = document.getElementById('confirmedTeamEnemyChart').getContext('2d');
            chartInstances.push(new Chart(confirmedTeamEnemyCtx, {
                type: 'doughnut',
                data: {
                    labels: ['Team Reports', 'Enemy Reports'],
                    datasets: [{
                        data: [totalConfirmedTeamReports, totalConfirmedEnemyReports],
                        backgroundColor: ['#22c55e', '#ef4444'],
                        borderColor: ['#16a34a', '#dc2626'],
                        borderWidth: 2
                    }]
                },
                options: {
                    ...chartOptions,
                    plugins: {
                        ...chartOptions.plugins,
                        title: {
                            display: true,
                            text: 'Confirmed: Team vs Enemy',
                            color: '#f1f5f9',
                            font: { size: 14, weight: 'bold' }
                        }
                    }
                }
            }));

            const confirmedStatusCtx = document.getElementById('confirmedStatusChart').getContext('2d');
            chartInstances.push(new Chart(confirmedStatusCtx, {
                type: 'bar',
                data: {
                    labels: ['Team', 'Enemy'],
                    datasets: [{
                        label: 'Confirmed Reports',
                        data: [totalConfirmedTeamReports, totalConfirmedEnemyReports],
                        backgroundColor: '#22c55e',
                        borderColor: '#16a34a',
                        borderWidth: 1
                    }]
                },
                options: {
                    ...chartOptions,
                    plugins: {
                        ...chartOptions.plugins,
                        title: {
                            display: true,
                            text: 'Confirmed Reports by Team',
                            color: '#f1f5f9',
                            font: { size: 14, weight: 'bold' }
                        }
                    },
                    scales: {
                        ...chartOptions.scales,
                        y: {
                            ...chartOptions.scales.y,
                            beginAtZero: true,
                            ticks: {
                                ...chartOptions.scales.y.ticks,
                                stepSize: 1,
                                precision: 0
                            }
                        }
                    }
                }
            }));

            if (matchData.length > 1) {
                const confirmedPerMatchCtx = document.getElementById('confirmedPerMatchChart').getContext('2d');
                const matchLabels = matchData.map(m => `Match ${m.matchID}`);
                chartInstances.push(new Chart(confirmedPerMatchCtx, {
                    type: 'bar',
                    data: {
                        labels: matchLabels,
                        datasets: [
                            {
                                label: 'Team Reports',
                                data: matchData.map(m => m.confirmedTeamReports),
                                backgroundColor: '#22c55e',
                                borderColor: '#16a34a',
                                borderWidth: 1
                            },
                            {
                                label: 'Enemy Reports',
                                data: matchData.map(m => m.confirmedEnemyReports),
                                backgroundColor: '#ef4444',
                                borderColor: '#dc2626',
                                borderWidth: 1
                            }
                        ]
                    },
                    options: {
                        ...chartOptions,
                        plugins: {
                            ...chartOptions.plugins,
                            title: {
                                display: true,
                                text: 'Confirmed Reports Per Match',
                                color: '#f1f5f9',
                                font: { size: 14, weight: 'bold' }
                            }
                        },
                        scales: {
                            ...chartOptions.scales,
                            y: {
                                ...chartOptions.scales.y,
                                beginAtZero: true,
                                ticks: {
                                    ...chartOptions.scales.y.ticks,
                                    stepSize: 1,
                                    precision: 0
                                }
                            }
                        }
                    }
                }));
                document.getElementById('confirmedPerMatchContainer').style.display = 'block';
            } else {
                document.getElementById('confirmedPerMatchContainer').style.display = 'none';
            }

            const confirmedTimelineCtx = document.getElementById('confirmedTimelineChart').getContext('2d');
            const confirmedFriendlyData = confirmedTimelineData.filter(d => d.y === 1);
            const confirmedEnemyData = confirmedTimelineData.filter(d => d.y === 2);
            
            chartInstances.push(new Chart(confirmedTimelineCtx, {
                type: 'scatter',
                data: {
                    datasets: [
                        {
                            label: 'Team Reports',
                            data: confirmedFriendlyData,
                            backgroundColor: '#22c55e80',
                            borderColor: '#22c55e',
                            borderWidth: 2,
                            pointRadius: 5
                        },
                        {
                            label: 'Enemy Reports',
                            data: confirmedEnemyData,
                            backgroundColor: '#ef444480',
                            borderColor: '#ef4444',
                            borderWidth: 2,
                            pointRadius: 5
                        }
                    ]
                },
                options: {
                    ...chartOptions,
                    plugins: {
                        ...chartOptions.plugins,
                        title: {
                            display: true,
                            text: 'Confirmed Reports Timeline (Match Minutes)',
                            color: '#f1f5f9',
                            font: { size: 14, weight: 'bold' }
                        },
                        tooltip: {
                            callbacks: {
                                label: function(context) {
                                    const point = context.raw;
                                    return `${context.dataset.label}: ${point.x.toFixed(2)} min (Confirmed)`;
                                }
                            }
                        }
                    },
                    scales: {
                        ...chartOptions.scales,
                        x: {
                            ...chartOptions.scales.x,
                            title: {
                                display: true,
                                text: 'Match Time (minutes)',
                                color: '#94a3b8'
                            }
                        },
                        y: {
                            ...chartOptions.scales.y,
                            min: 0.5,
                            max: 2.5,
                            ticks: {
                                stepSize: 1,
                                callback: function(value) {
                                    return value === 1 ? 'Team' : value === 2 ? 'Enemy' : '';
                                }
                            },
                            title: {
                                display: true,
                                text: 'Report Type',
                                color: '#94a3b8'
                            }
                        }
                    }
                }
            }));

            const confirmedTopReportersArray = Array.from(confirmedPlayerReportCounts.entries())
                .sort((a, b) => b[1] - a[1])
                .slice(0, 10);
            
            if (confirmedTopReportersArray.length > 0) {
                const confirmedTopReportersCtx = document.getElementById('confirmedTopReportersChart').getContext('2d');
                chartInstances.push(new Chart(confirmedTopReportersCtx, {
                    type: 'bar',
                    data: {
                        labels: confirmedTopReportersArray.map(p => p[0]),
                        datasets: [{
                            label: 'Confirmed Report Count',
                            data: confirmedTopReportersArray.map(p => p[1]),
                            backgroundColor: '#22c55e',
                            borderColor: '#16a34a',
                            borderWidth: 1
                        }]
                    },
                    options: {
                        ...chartOptions,
                        indexAxis: 'y',
                        plugins: {
                            ...chartOptions.plugins,
                            title: {
                                display: true,
                                text: 'Top Reporters (Confirmed)',
                                color: '#f1f5f9',
                                font: { size: 14, weight: 'bold' }
                            }
                        },
                        scales: {
                            ...chartOptions.scales,
                            x: {
                                ...chartOptions.scales.x,
                                beginAtZero: true,
                                ticks: {
                                    ...chartOptions.scales.x.ticks,
                                    stepSize: 1,
                                    precision: 0
                                }
                            }
                        }
                    }
                }));
            }
        }

        if (totalUnconfirmedTeamReports + totalUnconfirmedEnemyReports > 0) {
            const unconfirmedTeamEnemyCtx = document.getElementById('unconfirmedTeamEnemyChart').getContext('2d');
            chartInstances.push(new Chart(unconfirmedTeamEnemyCtx, {
                type: 'doughnut',
                data: {
                    labels: ['Team Reports', 'Enemy Reports'],
                    datasets: [{
                        data: [totalUnconfirmedTeamReports, totalUnconfirmedEnemyReports],
                        backgroundColor: ['#22c55e', '#ef4444'],
                        borderColor: ['#16a34a', '#dc2626'],
                        borderWidth: 2
                    }]
                },
                options: {
                    ...chartOptions,
                    plugins: {
                        ...chartOptions.plugins,
                        title: {
                            display: true,
                            text: 'Unconfirmed: Team vs Enemy',
                            color: '#f1f5f9',
                            font: { size: 14, weight: 'bold' }
                        }
                    }
                }
            }));

            const unconfirmedStatusCtx = document.getElementById('unconfirmedStatusChart').getContext('2d');
            chartInstances.push(new Chart(unconfirmedStatusCtx, {
                type: 'bar',
                data: {
                    labels: ['Team', 'Enemy'],
                    datasets: [{
                        label: 'Unconfirmed Reports',
                        data: [totalUnconfirmedTeamReports, totalUnconfirmedEnemyReports],
                        backgroundColor: '#eab308',
                        borderColor: '#ca8a04',
                        borderWidth: 1
                    }]
                },
                options: {
                    ...chartOptions,
                    plugins: {
                        ...chartOptions.plugins,
                        title: {
                            display: true,
                            text: 'Unconfirmed Reports by Team',
                            color: '#f1f5f9',
                            font: { size: 14, weight: 'bold' }
                        }
                    },
                    scales: {
                        ...chartOptions.scales,
                        y: {
                            ...chartOptions.scales.y,
                            beginAtZero: true
                        }
                    }
                }
            }));

            if (matchData.length > 1) {
                const unconfirmedPerMatchCtx = document.getElementById('unconfirmedPerMatchChart').getContext('2d');
                const matchLabels = matchData.map(m => `Match ${m.matchID}`);
                chartInstances.push(new Chart(unconfirmedPerMatchCtx, {
                    type: 'bar',
                    data: {
                        labels: matchLabels,
                        datasets: [
                            {
                                label: 'Team Reports',
                                data: matchData.map(m => m.unconfirmedTeamReports),
                                backgroundColor: '#22c55e',
                                borderColor: '#16a34a',
                                borderWidth: 1
                            },
                            {
                                label: 'Enemy Reports',
                                data: matchData.map(m => m.unconfirmedEnemyReports),
                                backgroundColor: '#ef4444',
                                borderColor: '#dc2626',
                                borderWidth: 1
                            }
                        ]
                    },
                    options: {
                        ...chartOptions,
                        plugins: {
                            ...chartOptions.plugins,
                            title: {
                                display: true,
                                text: 'Unconfirmed Reports Per Match',
                                color: '#f1f5f9',
                                font: { size: 14, weight: 'bold' }
                            }
                        },
                        scales: {
                            ...chartOptions.scales,
                            y: {
                                ...chartOptions.scales.y,
                                beginAtZero: true,
                                ticks: {
                                    ...chartOptions.scales.y.ticks,
                                    stepSize: 1,
                                    precision: 0
                                }
                            }
                        }
                    }
                }));
                document.getElementById('unconfirmedPerMatchContainer').style.display = 'block';
            } else {
                document.getElementById('unconfirmedPerMatchContainer').style.display = 'none';
            }

            const unconfirmedTimelineCtx = document.getElementById('unconfirmedTimelineChart').getContext('2d');
            const unconfirmedFriendlyData = unconfirmedTimelineData.filter(d => d.y === 1);
            const unconfirmedEnemyData = unconfirmedTimelineData.filter(d => d.y === 2);
            
            chartInstances.push(new Chart(unconfirmedTimelineCtx, {
                type: 'scatter',
                data: {
                    datasets: [
                        {
                            label: 'Team Reports',
                            data: unconfirmedFriendlyData,
                            backgroundColor: '#22c55e80',
                            borderColor: '#22c55e',
                            borderWidth: 2,
                            pointRadius: 5
                        },
                        {
                            label: 'Enemy Reports',
                            data: unconfirmedEnemyData,
                            backgroundColor: '#ef444480',
                            borderColor: '#ef4444',
                            borderWidth: 2,
                            pointRadius: 5
                        }
                    ]
                },
                options: {
                    ...chartOptions,
                    plugins: {
                        ...chartOptions.plugins,
                        title: {
                            display: true,
                            text: 'Unconfirmed Reports Timeline (Match Minutes)',
                            color: '#f1f5f9',
                            font: { size: 14, weight: 'bold' }
                        },
                        tooltip: {
                            callbacks: {
                                label: function(context) {
                                    const point = context.raw;
                                    return `${context.dataset.label}: ${point.x.toFixed(2)} min (Unconfirmed)`;
                                }
                            }
                        }
                    },
                    scales: {
                        ...chartOptions.scales,
                        x: {
                            ...chartOptions.scales.x,
                            title: {
                                display: true,
                                text: 'Match Time (minutes)',
                                color: '#94a3b8'
                            }
                        },
                        y: {
                            ...chartOptions.scales.y,
                            min: 0.5,
                            max: 2.5,
                            ticks: {
                                stepSize: 1,
                                callback: function(value) {
                                    return value === 1 ? 'Team' : value === 2 ? 'Enemy' : '';
                                }
                            },
                            title: {
                                display: true,
                                text: 'Report Type',
                                color: '#94a3b8'
                            }
                        }
                    }
                }
            }));

            const unconfirmedTopReportersArray = Array.from(unconfirmedPlayerReportCounts.entries())
                .sort((a, b) => b[1] - a[1])
                .slice(0, 10);
            
            if (unconfirmedTopReportersArray.length > 0) {
                const unconfirmedTopReportersCtx = document.getElementById('unconfirmedTopReportersChart').getContext('2d');
                chartInstances.push(new Chart(unconfirmedTopReportersCtx, {
                    type: 'bar',
                    data: {
                        labels: unconfirmedTopReportersArray.map(p => p[0]),
                        datasets: [{
                            label: 'Unconfirmed Report Count',
                            data: unconfirmedTopReportersArray.map(p => p[1]),
                            backgroundColor: '#eab308',
                            borderColor: '#ca8a04',
                            borderWidth: 1
                        }]
                    },
                    options: {
                        ...chartOptions,
                        indexAxis: 'y',
                        plugins: {
                            ...chartOptions.plugins,
                            title: {
                                display: true,
                                text: 'Top Reporters (Unconfirmed)',
                                color: '#f1f5f9',
                                font: { size: 14, weight: 'bold' }
                            }
                        },
                        scales: {
                            ...chartOptions.scales,
                            x: {
                                ...chartOptions.scales.x,
                                beginAtZero: true
                            }
                        }
                    }
                }));
            }
        }
    }
});
