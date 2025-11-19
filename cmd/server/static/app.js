document.addEventListener('DOMContentLoaded', () => {
    const replayDirInput = document.getElementById('replay-dir');
    const stratzApiTokenInput = document.getElementById('stratz-api-token');
    const steamApiKeyInput = document.getElementById('steam-api-key');
    const saveConfigBtn = document.getElementById('save-config');
    const replayList = document.getElementById('replay-list');
    const refreshReplaysBtn = document.getElementById('refresh-replays');
    const sortDateBtn = document.getElementById('sort-date');
    const selectAllBtn = document.getElementById('select-all');
    const deselectAllBtn = document.getElementById('deselect-all');
    const deleteSelectedBtn = document.getElementById('delete-selected');
    const startParseBtn = document.getElementById('start-parse');
    const steamIdInput = document.getElementById('steam-id');
    const slotIdInput = document.getElementById('slot-id');
    const progressSection = document.getElementById('progress-section');
    const progressBar = document.getElementById('progress-bar');
    const progressText = document.getElementById('progress-text');
    const resultsSection = document.getElementById('results-section');
    const conclusionDiv = document.getElementById('conclusion');
    const resultsLog = document.getElementById('results-log');
    const graphsSection = document.getElementById('graphs-section');
    const historySteamIdInput = document.getElementById('history-steam-id');
    const historyLimitInput = document.getElementById('history-limit');
    const fetchHistoryBtn = document.getElementById('fetch-history');
    const historyResults = document.getElementById('history-results');

    // Profile Elements
    const profileSelect = document.getElementById('profile-select');
    const newProfileName = document.getElementById('new-profile-name');
    const newProfileId = document.getElementById('new-profile-id');
    const addProfileBtn = document.getElementById('add-profile-btn');
    const deleteProfileBtn = document.getElementById('delete-profile-btn');

    // Profile Logic
    let profiles = [];

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
        renderProfileSelect();
    }

    function saveProfiles() {
        localStorage.setItem('steamProfiles', JSON.stringify(profiles));
        renderProfileSelect();
    }

    function renderProfileSelect() {
        // Keep selected value if possible
        const currentVal = profileSelect.value;
        
        profileSelect.innerHTML = '<option value="">Select a profile...</option>';
        profiles.forEach((p, index) => {
            const option = document.createElement('option');
            option.value = index;
            option.textContent = `${p.name} (${p.id})`;
            profileSelect.appendChild(option);
        });

        if (currentVal && profiles[currentVal]) {
            profileSelect.value = currentVal;
            deleteProfileBtn.style.visibility = 'visible';
        } else {
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
        const index = profileSelect.value;
        if (index !== "") {
            const p = profiles[index];
            // Fill Steam ID fields
            if (historySteamIdInput) {
                historySteamIdInput.value = p.id;
                historySteamIdInput.dispatchEvent(new Event('input')); 
            }
            if (steamIdInput) {
                steamIdInput.value = p.id;
                steamIdInput.dispatchEvent(new Event('input'));
            }
            deleteProfileBtn.style.visibility = 'visible';
            loadReplays();
        } else {
            deleteProfileBtn.style.visibility = 'hidden';
            loadReplays();
        }
    });

    function getSelectedProfileName() {
        const index = profileSelect.value;
        if (index !== "" && profiles[index]) {
            return profiles[index].name;
        }
        return "";
    }

    // Initialize Profiles
    loadProfiles();
    
    // Steam Login Elements
    const steamUserInput = document.getElementById('steam-user');
    const steamPassInput = document.getElementById('steam-pass');
    const steamCodeInput = document.getElementById('steam-code');
    const steamGuardGroup = document.getElementById('steam-guard-group');
    const steamLoginBtn = document.getElementById('steam-login-btn');
    const steamStatusText = document.getElementById('steam-status');

    // Steam Logic
    let steamPollingInterval;

    function updateSteamUI(status, text) {
        steamStatusText.textContent = `Status: ${text}`;
        // StatusNeedGuardCode = 2
        if (status === 2) {
            steamStatusText.style.color = '#ff9800';
            steamGuardGroup.classList.remove('hidden');
            steamLoginBtn.textContent = 'Submit Code';
            steamLoginBtn.disabled = false;
        } else if (status === 4) { // GCReady
             steamStatusText.style.color = '#4caf50';
             steamLoginBtn.textContent = 'Connected';
             steamLoginBtn.disabled = true;
             steamGuardGroup.classList.add('hidden');
             clearInterval(steamPollingInterval); // Stop polling when ready (optional, or keep checking disconnect)
             // Actually better to keep polling slowly to detect disconnects
             if (steamPollingInterval) clearInterval(steamPollingInterval);
             steamPollingInterval = setInterval(pollSteamStatus, 10000);
        } else {
            steamStatusText.style.color = '';
            steamGuardGroup.classList.add('hidden');
            steamLoginBtn.textContent = 'Connect to Steam';
            steamLoginBtn.disabled = false;
        }
    }

    function pollSteamStatus() {
        fetch('/api/steam/status')
            .then(res => res.json())
            .then(data => {
                updateSteamUI(data.status, data.statusText);
            })
            .catch(err => console.error('Steam status poll failed:', err));
    }

    // Start polling on load
    pollSteamStatus();
    steamPollingInterval = setInterval(pollSteamStatus, 2000);

    steamLoginBtn.addEventListener('click', () => {
        const username = steamUserInput.value.trim();
        const password = steamPassInput.value.trim();
        const code = steamCodeInput.value.trim();

        if (!username || !password) {
            alert('Please enter both username and password');
            return;
        }

        if (steamLoginBtn.textContent === 'Submit Code' && !code) {
            alert('Please enter your Steam Guard code');
            return;
        }

        const originalText = steamLoginBtn.textContent;
        steamLoginBtn.disabled = true;
        steamLoginBtn.textContent = code ? 'Submitting Code...' : 'Connecting...';

        fetch('/api/steam/login', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username, password, code })
        })
        .then(res => {
            if (!res.ok) {
                return res.text().then(text => {
                    throw new Error(`Server error (${res.status}): ${text}`);
                });
            }
            return res.json();
        })
        .then(data => {
             // Status update will happen via polling
             pollSteamStatus();
             // Re-enable button after a short delay if status hasn't changed to ready
             setTimeout(() => {
                 const currentStatus = parseInt(steamStatusText.textContent.match(/Status: (.+)/)?.[1] || '');
                 if (steamLoginBtn.textContent.includes('Submitting') || steamLoginBtn.textContent.includes('Connecting')) {
                     if (data.status !== 4) { // Not GCReady
                         steamLoginBtn.disabled = false;
                         steamLoginBtn.textContent = originalText;
                     }
                 }
             }, 2000);
        })
        .catch(err => {
            console.error('Steam login error:', err);
            alert('Failed to send login request: ' + err.message);
            steamLoginBtn.disabled = false;
            steamLoginBtn.textContent = originalText;
        });
    });
    
    // Load from localStorage
    function loadFromStorage() {
        const savedToken = localStorage.getItem('stratzApiToken');
        const savedSteamKey = localStorage.getItem('steamApiKey');
        const savedSteamUser = localStorage.getItem('steamUser');
        const savedSteamPass = localStorage.getItem('steamPass');
        const savedSteamId = localStorage.getItem('steamId');
        const savedSlotId = localStorage.getItem('slotId');
        
        if (savedToken) {
            stratzApiTokenInput.value = savedToken;
            console.log('Loaded token from localStorage (length:', savedToken.length, ')');
        }
        if (savedSteamKey) {
            steamApiKeyInput.value = savedSteamKey;
        }
        if (savedSteamUser) {
            steamUserInput.value = savedSteamUser;
        }
        if (savedSteamPass) {
            steamPassInput.value = savedSteamPass;
        }
        if (savedSteamId) steamIdInput.value = savedSteamId;
        if (savedSlotId) slotIdInput.value = savedSlotId;
    }

    // Save to localStorage
    function saveToStorage() {
        if (stratzApiTokenInput.value) localStorage.setItem('stratzApiToken', stratzApiTokenInput.value);
        if (steamApiKeyInput.value) localStorage.setItem('steamApiKey', steamApiKeyInput.value);
        if (steamUserInput.value) localStorage.setItem('steamUser', steamUserInput.value);
        if (steamPassInput.value) localStorage.setItem('steamPass', steamPassInput.value);
        if (steamIdInput.value) localStorage.setItem('steamId', steamIdInput.value);
        if (slotIdInput.value) localStorage.setItem('slotId', slotIdInput.value);
    }

    // Load from localStorage first (before server config)
    loadFromStorage();

    // Auto-save token to server if it exists in localStorage on page load
    const initialToken = localStorage.getItem('stratzApiToken');
    if (initialToken && initialToken.trim() !== '') {
        setTimeout(() => {
            fetch('/api/config', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ stratzApiToken: initialToken })
            }).catch(err => console.error('Auto-save on load failed:', err));
        }, 500);
    }

    // Final safety check after a short delay to catch any browser clearing behavior
    setTimeout(() => {
        if (!stratzApiTokenInput.value) {
            const savedToken = localStorage.getItem('stratzApiToken');
            if (savedToken) {
                console.log('Late check: Token field was cleared, restoring from localStorage');
                stratzApiTokenInput.value = savedToken;
            }
        }
    }, 1000);

    // Load Config from server (will override localStorage if present)
    fetch('/api/config')
        .then(res => res.json())
        .then(config => {
            replayDirInput.value = config.replayDir || '';
            // Only update token if server has a non-empty value
            if (config.stratzApiToken && config.stratzApiToken.trim() !== '') {
                console.log('Server has token, updating from server (length:', config.stratzApiToken.length, ')');
                stratzApiTokenInput.value = config.stratzApiToken;
                localStorage.setItem('stratzApiToken', config.stratzApiToken);
            } else {
                console.log('Server token is empty/missing, keeping localStorage value');
            }
            
            if (config.steamApiKey && config.steamApiKey.trim() !== '') {
                steamApiKeyInput.value = config.steamApiKey;
                localStorage.setItem('steamApiKey', config.steamApiKey);
            }

            if (config.steamUser && config.steamUser.trim() !== '') {
                steamUserInput.value = config.steamUser;
                localStorage.setItem('steamUser', config.steamUser);
            }
            
            if (config.steamPass && config.steamPass.trim() !== '') {
                steamPassInput.value = config.steamPass;
                localStorage.setItem('steamPass', config.steamPass);
            }

            // If server token is empty/missing, keep the localStorage value (already loaded above)
            // Double-check: if token field is empty but localStorage has it, restore it
            if (!stratzApiTokenInput.value) {
                const savedToken = localStorage.getItem('stratzApiToken');
                if (savedToken) {
                    console.log('Token field was cleared, restoring from localStorage');
                    stratzApiTokenInput.value = savedToken;
                }
            }
            loadReplays();
        })
        .catch(() => {
            // If server config fails, values from loadFromStorage() above will be used
            // Double-check token is still there
            if (!stratzApiTokenInput.value) {
                const savedToken = localStorage.getItem('stratzApiToken');
                if (savedToken) {
                    stratzApiTokenInput.value = savedToken;
                }
            }
        });

    // Auto-save token when it changes (with debounce)
    let saveTokenTimeout;
    const autoSave = () => {
        const token = stratzApiTokenInput.value.trim();
        const steamKey = steamApiKeyInput.value.trim();
        // Also save steam credentials locally if changed
        const steamUser = steamUserInput.value.trim();
        const steamPass = steamPassInput.value.trim();
        
        localStorage.setItem('stratzApiToken', token);
        localStorage.setItem('steamApiKey', steamKey);
        if (steamUser) localStorage.setItem('steamUser', steamUser);
        if (steamPass) localStorage.setItem('steamPass', steamPass);
        
        // Auto-save to server after user stops typing (1 second delay)
        clearTimeout(saveTokenTimeout);
        saveTokenTimeout = setTimeout(() => {
            const payload = {};
            if (token) payload.stratzApiToken = token;
            if (steamKey) payload.steamApiKey = steamKey;
            if (steamUser) payload.steamUser = steamUser;
            if (steamPass) payload.steamPass = steamPass;

            if (Object.keys(payload).length > 0) {
                fetch('/api/config', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(payload)
                }).catch(err => console.error('Auto-save failed:', err));
            }
        }, 1000);
    };

    stratzApiTokenInput.addEventListener('input', autoSave);
    steamApiKeyInput.addEventListener('input', autoSave);
    stratzApiTokenInput.addEventListener('change', autoSave);
    steamApiKeyInput.addEventListener('change', autoSave);
    steamUserInput.addEventListener('input', autoSave);
    steamPassInput.addEventListener('input', autoSave);
    
    steamIdInput.addEventListener('input', () => {
        if (steamIdInput.value) localStorage.setItem('steamId', steamIdInput.value);
    });
    slotIdInput.addEventListener('input', () => {
        if (slotIdInput.value) localStorage.setItem('slotId', slotIdInput.value);
    });

    // Save Config
    saveConfigBtn.addEventListener('click', () => {
        const newDir = replayDirInput.value;
        const newToken = stratzApiTokenInput.value.trim();
        const newSteamKey = steamApiKeyInput.value.trim();
        const newSteamUser = steamUserInput.value.trim();
        const newSteamPass = steamPassInput.value.trim();

        console.log('Saving config - Token length:', newToken.length);
        
        // Save to localStorage immediately
        if (newToken) {
            localStorage.setItem('stratzApiToken', newToken);
            console.log('Saved token to localStorage');
        } else {
            console.warn('Token is empty when saving!');
        }
        if (newSteamKey) {
            localStorage.setItem('steamApiKey', newSteamKey);
        }
        if (newSteamUser) localStorage.setItem('steamUser', newSteamUser);
        if (newSteamPass) localStorage.setItem('steamPass', newSteamPass);

        saveToStorage();
        
        const payload = { replayDir: newDir };
        if (newToken) {
            payload.stratzApiToken = newToken;
        }
        if (newSteamKey) {
            payload.steamApiKey = newSteamKey;
        }
        if (newSteamUser) payload.steamUser = newSteamUser;
        if (newSteamPass) payload.steamPass = newSteamPass;
        
        fetch('/api/config', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload)
        })
            .then(res => res.json())
            .then(config => {
                console.log('Server response - Token length:', config.stratzApiToken ? config.stratzApiToken.length : 0);
                // Ensure localStorage is updated with what server saved
                if (config.stratzApiToken) {
                    localStorage.setItem('stratzApiToken', config.stratzApiToken);
                    stratzApiTokenInput.value = config.stratzApiToken;
                }
                if (config.steamApiKey) {
                    localStorage.setItem('steamApiKey', config.steamApiKey);
                    steamApiKeyInput.value = config.steamApiKey;
                }
                if (config.steamUser) {
                    localStorage.setItem('steamUser', config.steamUser);
                    steamUserInput.value = config.steamUser;
                }
                if (config.steamPass) {
                    localStorage.setItem('steamPass', config.steamPass);
                    steamPassInput.value = config.steamPass;
                }
                alert('Configuration saved!');
                loadReplays();
            })
            .catch(err => {
                console.error('Error saving config:', err);
                alert('Error saving config: ' + err);
            });
    });

    let currentReplays = [];
    let sortOrder = 'desc'; // 'desc' (newest first) or 'asc' (oldest first)

    // Load Replays
    function loadReplays() {
        replayList.innerHTML = '<p class="loading">Loading replays...</p>';
        const profileName = getSelectedProfileName();
        const url = '/api/replays?t=' + Date.now() + (profileName ? '&profile=' + encodeURIComponent(profileName) : '');
        fetch(url)
            .then(res => res.json())
            .then(replays => {
                currentReplays = replays || [];
                renderReplayList();
                if (window.updateHistoryStatus) {
                    window.updateHistoryStatus();
                }
            })
            .catch(err => {
                replayList.innerHTML = `<p style="color:red">Error loading replays: ${err}</p>`;
            });
    }

    setInterval(() => {
        if (document.visibilityState === 'visible') {
            loadReplays();
        }
    }, 120000);
    
    function renderReplayList() {
        replayList.innerHTML = '';
        if (!currentReplays || currentReplays.length === 0) {
            replayList.innerHTML = '<p>No .dem files found in directory.</p>';
            return;
        }

        // Sort replays
        currentReplays.sort((a, b) => {
            const dateA = a.date ? new Date(a.date) : new Date(0);
            const dateB = b.date ? new Date(b.date) : new Date(0);
            return sortOrder === 'desc' ? dateB - dateA : dateA - dateB;
        });

        currentReplays.forEach(replay => {
            const div = document.createElement('div');
            div.className = 'replay-item';
            const fileName = replay.fileName || replay;
            const id = fileName.replace('.dem', '');
            const date = replay.date ? new Date(replay.date).toLocaleString() : '';
            const dateDisplay = date ? ` <span class="date-display">(${date})</span>` : '';
            div.innerHTML = `
                <input type="checkbox" value="${id}" id="replay-${id}">
                <label for="replay-${id}">${fileName}${dateDisplay}</label>
            `;
            replayList.appendChild(div);
        });
    }

    sortDateBtn.addEventListener('click', () => {
        sortOrder = sortOrder === 'desc' ? 'asc' : 'desc';
        sortDateBtn.textContent = sortOrder === 'desc' ? 'Sort: Newest' : 'Sort: Oldest';
        renderReplayList();
    });
    
    window.loadReplays = loadReplays;

    function deleteSelectedReplays() {
        const checkboxes = document.querySelectorAll('.replay-item input[type="checkbox"]:checked');
        const selectedIds = Array.from(checkboxes).map(cb => cb.value);
        
        if (selectedIds.length === 0) {
            alert('Please select at least one replay to delete.');
            return;
        }
        
        const fileNames = selectedIds.map(id => id + '.dem').join(', ');
        if (!confirm(`Are you sure you want to delete ${selectedIds.length} replay file(s)?\n\n${fileNames}\n\nThis cannot be undone.`)) {
            return;
        }
        
        const originalText = deleteSelectedBtn.textContent;
        deleteSelectedBtn.disabled = true;
        deleteSelectedBtn.textContent = 'Deleting...';
        
        const deletePromises = selectedIds.map(matchId => 
            fetch('/api/delete', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ 
                    matchId: matchId,
                    profileName: getSelectedProfileName()
                })
            }).then(async res => {
                if (!res.ok) {
                    const errorText = await res.text();
                    throw new Error(`Failed to delete ${matchId}: ${errorText || res.statusText}`);
                }
                return res.json();
            })
        );
        
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
                loadReplays();
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

    refreshReplaysBtn.addEventListener('click', loadReplays);

    selectAllBtn.addEventListener('click', () => {
        document.querySelectorAll('.replay-item input[type="checkbox"]').forEach(cb => cb.checked = true);
    });

    deselectAllBtn.addEventListener('click', () => {
        document.querySelectorAll('.replay-item input[type="checkbox"]').forEach(cb => cb.checked = false);
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

        // Use string for Steam ID to avoid precision loss, convert to number for JSON
        const steamIdStr = steamIdInput.value ? steamIdInput.value.trim() : '';
        // Keep as string
        const steamIdToSend = steamIdStr || "0";
        const slotId = slotIdInput.value ? parseInt(slotIdInput.value) : -1;
        saveToStorage();

        // Reset UI
        progressSection.classList.remove('hidden');
        resultsSection.classList.add('hidden');
        resultsLog.innerHTML = '';
        conclusionDiv.innerHTML = '';
        graphsSection.classList.add('hidden');
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

        for (const matchId of selectedIds) {
            progressText.textContent = `Processing ${processedCount + 1} / ${selectedIds.length}: ${matchId}`;

            try {
                const res = await fetch('/api/parse', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        matchId: matchId,
                        reportedSlot: slotId,
                        reportedSteamId: steamIdToSend,
                        profileName: getSelectedProfileName()
                    })
                });

                if (!res.ok) throw new Error(await res.text());

                const result = await res.json();

                // Log results
                const logEntry = document.createElement('div');
                logEntry.className = 'log-entry';
                
                let headerHtml = `<div class="log-header"><strong>Match ${result.MatchID}</strong></div>`;
                headerHtml += `<div class="log-summary">Reports: ${result.TeamReports} (Friendly) / ${result.EnemyReports} (Enemy)</div>`;
                headerHtml += `<div class="log-summary">Confirmed: ${result.ConfirmedTeamReports || 0} (Friendly) / ${result.ConfirmedEnemyReports || 0} (Enemy)</div>`;
                
                let reportsHtml = '';
                if (result.Reports) {
                    reportsHtml = '<div class="log-reports">';
                    result.Reports.forEach(report => {
                        let reportText = `_REPORT_ (${report.Time}) from _${report.Team}_: ${report.Name} (Slot ${report.Slot})`;
                        let className = "report-item";
                        if (report.Confirmed) {
                            reportText += ` [CONFIRMED +${report.ConfirmationDelayMs}ms]`;
                            className += " confirmed";
                        } else {
                            reportText += ` [UNCONFIRMED]`;
                        }
                        reportsHtml += `<div class="${className}">${reportText}</div>`;
                    });
                    reportsHtml += '</div>';
                }
                logEntry.innerHTML = headerHtml + reportsHtml;
                resultsLog.appendChild(logEntry);

                totalTeamReports += result.TeamReports;
                totalEnemyReports += result.EnemyReports;
                totalConfirmedTeamReports += (result.ConfirmedTeamReports || 0);
                totalConfirmedEnemyReports += (result.ConfirmedEnemyReports || 0);

                const confirmedTeamReports = result.ConfirmedTeamReports || 0;
                const confirmedEnemyReports = result.ConfirmedEnemyReports || 0;
                const unconfirmedTeamReports = result.TeamReports - confirmedTeamReports;
                const unconfirmedEnemyReports = result.EnemyReports - confirmedEnemyReports;

                matchData.push({
                    matchID: result.MatchID,
                    teamReports: result.TeamReports,
                    enemyReports: result.EnemyReports,
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
                const errorEntry = document.createElement('div');
                errorEntry.style.color = 'red';
                errorEntry.className = 'log-entry';
                let errorMsg = err.message;
                if (errorMsg.includes('no such file or directory') || errorMsg.includes('Could not open replay file')) {
                    errorMsg = `Replay file not found: ${matchId}.dem - Make sure the replay file exists in your replay directory. You may need to download it from Dota 2 first.`;
                }
                errorEntry.innerHTML = `<strong style="color:red">Error processing ${matchId}:</strong> ${errorMsg}`;
                resultsLog.appendChild(errorEntry);
            }

            processedCount++;
            progressBar.style.width = `${(processedCount / selectedIds.length) * 100}%`;
        }

        // Conclusion
        resultsSection.classList.remove('hidden');
        conclusionDiv.innerHTML = `
            <h3>Analysis Complete</h3>
            <p>Processed ${processedCount} replays.</p>
            <p><strong>Total Team Reports:</strong> ${totalTeamReports} (Confirmed: ${totalConfirmedTeamReports})</p>
            <p><strong>Total Enemy Reports:</strong> ${totalEnemyReports} (Confirmed: ${totalConfirmedEnemyReports})</p>
            <p><em>${totalTeamReports + totalEnemyReports > 0 ? "You have been naughty!" : "Clean record! Well played."}</em></p>
        `;

        if (totalTeamReports + totalEnemyReports > 0) {
            const totalUnconfirmedTeamReports = totalTeamReports - totalConfirmedTeamReports;
            const totalUnconfirmedEnemyReports = totalEnemyReports - totalConfirmedEnemyReports;
            generateGraphs(matchData, confirmedPlayerReportCounts, unconfirmedPlayerReportCounts, 
                confirmedTimelineData, unconfirmedTimelineData, 
                totalTeamReports, totalEnemyReports, 
                totalConfirmedTeamReports, totalConfirmedEnemyReports,
                totalUnconfirmedTeamReports, totalUnconfirmedEnemyReports);
            graphsSection.classList.remove('hidden');
        }
    });

    let chartInstances = [];

    function generateGraphs(matchData, confirmedPlayerReportCounts, unconfirmedPlayerReportCounts, 
        confirmedTimelineData, unconfirmedTimelineData,
        totalTeamReports, totalEnemyReports, 
        totalConfirmedTeamReports, totalConfirmedEnemyReports,
        totalUnconfirmedTeamReports, totalUnconfirmedEnemyReports) {
        chartInstances.forEach(chart => chart.destroy());
        chartInstances = [];

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
                            beginAtZero: true
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
                                beginAtZero: true
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
                                beginAtZero: true
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
                                beginAtZero: true
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
