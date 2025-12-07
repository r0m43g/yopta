<template lang="pug">
.flex.flex-col.items-center.justify-center.h-screen.w-screen.overflow-auto.scrollbar-thin
  .card.bg-base-200(class="w-11/12 h-11/12 shadow-xl overflow-auto scrollbar-thin")
    .card-body
      h1.card-title.text-3xl.mb-6 Sistemos nustatymai
      .grid.grid-cols-2.gap-6
        div
          .card.bg-base-100.shadow-md
            .card-body.p-4
              h2.card-title.text-xl.mb-4
                svg.h-6.w-6.mr-2(xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor")
                  path(stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z")
                  path(stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z")
                | Sistemos nustatymai

              .overflow-x-auto
                table.table.table-zebra
                  thead
                    tr
                      th Nustatymas
                      th Reikšmė
                      th Veiksmai
                  tbody
                    tr(v-for="setting in settings" :key="setting.id" class="hover")
                      td
                        .font-bold {{ getDisplayName(setting.setting_key) }}
                        .text-xs.opacity-70 {{ setting.description }}
                      td
                        .form-control(v-if="isBooleanSetting(setting.setting_key)")
                          label.cursor-pointer.label
                            input.toggle.toggle-primary(
                              type="checkbox"
                              :checked="setting.setting_value === 'true'"
                              @change="updateBooleanSetting(setting, $event)"
                            )
                        div(v-else) {{ setting.setting_value }}
                      td
                        .flex.space-x-2
                          button.btn.btn-square.btn-sm.btn-ghost(@click="editSetting(setting)")
                            svg.h-5.w-5(xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor")
                              path(stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z")

        div
          .card.bg-base-100.shadow-md
            .card-body.p-4
              h2.card-title.text-xl.mb-4
                svg.h-6.w-6.mr-2(xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor")
                  path(stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z")
                | IP juodasis sąrašas

              .flex.mb-4
                input.input.input-bordered.w-full.max-w-xs.mr-2(
                  type="text"
                  v-model="newIP.ip_address"
                  placeholder="IP adresas"
                )
                input.input.input-bordered.w-full.max-w-xs.mr-2(
                  type="text"
                  v-model="newIP.reason"
                  placeholder="Priežastis"
                )
                button.btn.btn-primary(@click="addBlacklistedIP") Pridėti

              .overflow-x-auto
                table.table.table-zebra
                  thead
                    tr
                      th IP adresas
                      th Priežastis
                      th Pridėjo
                      th Data
                      th Veiksmai
                  tbody
                    tr(v-for="ip in blacklistedIPs" :key="ip.id" class="hover")
                      td {{ ip.ip_address }}
                      td {{ ip.reason }}
                      td {{ ip.username || 'Nežinomas' }}
                      td {{ formatDate(ip.created_at) }}
                      td
                        button.btn.btn-square.btn-sm.btn-ghost.text-error(@click="confirmDeleteIP(ip)")
                          svg.h-5.w-5(xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor")
                            path(stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16")

  .modal(:class="{'modal-open': editingSettings}")
    .modal-box
      h3.font-bold.text-lg Redaguoti nustatymą
      p.py-2 Redaguokite {{ getDisplayName(currentSetting?.setting_key) || '' }} nustatymą:

      .form-control.w-full.py-4(v-if="currentSetting")
        label.label
          span.label-text Reikšmė:
        div(v-if="isBooleanSetting(currentSetting.setting_key)")
          label.cursor-pointer.label.justify-start
            input.checkbox.checkbox-primary.mr-4(
              type="checkbox"
              v-model="editedSettingValue"
            )
            span.label-text Įgalinti
        input.input.input-bordered.w-full(
          v-else
          type="text"
          v-model="editedSettingValue"
          placeholder="Reikšmė"
        )

      .modal-action
        button.btn(@click="editingSettings = false") Atšaukti
        button.btn.btn-primary(@click="saveSetting") Išsaugoti

  .modal(:class="{'modal-open': ipToDelete}")
    .modal-box
      h3.font-bold.text-lg Patvirtinkite ištrynimą
      p.py-4
        | Ar tikrai norite pašalinti IP adresą
        span.font-bold {{ ipToDelete?.ip_address }}
        | iš juodojo sąrašo?
      .modal-action
        button.btn(@click="ipToDelete = null") Atšaukti
        button.btn.btn-error(@click="removeBlacklistedIP") Pašalinti
</template>

<script setup>
import { ref, onMounted } from 'vue';
import api from '@/services/api';
import { useLoggingStore } from '@/stores/logging';
import { useSlogStore } from '@/stores/slog';

// Store instances
const loggingStore = useLoggingStore();
const slogStore = useSlogStore();

// State variables
const settings = ref([]);
const blacklistedIPs = ref([]);
const newIP = ref({
  ip_address: '',
  reason: ''
});

// Modal state
const editingSettings = ref(false);
const currentSetting = ref(null);
const editedSettingValue = ref('');
const ipToDelete = ref(null);

/**
 * Maps setting keys to human-readable display names
 * @param {string} key - Setting key
 * @returns {string} Human-readable name
 */
const getDisplayName = (key) => {
  const names = {
    'registration_enabled': 'Registracija įgalinta'
  };
  return names[key] || key;
};

/**
 * Determines if a setting should be displayed as a boolean toggle
 * @param {string} key - Setting key
 * @returns {boolean} True if setting is boolean type
 */
const isBooleanSetting = (key) => {
  return ['registration_enabled'].includes(key);
};

/**
 * Formats a date string for display
 * @param {string} dateString - ISO date string
 * @returns {string} Formatted date
 */
const formatDate = (dateString) => {
  if (!dateString) return '';
  return new Date(dateString).toLocaleString('lt-LT');
};

/**
 * Loads system settings from the server
 */
const loadSettings = async () => {
  try {
    const response = await api.get('/system-settings');
    settings.value = response.data;

    loggingStore.info('Sistemos nustatymai užkrauti', {
      component: 'SystemSettings',
      action: 'load_settings',
      count: response.data.length
    });
  } catch (error) {
    slogStore.addToast({
      message: 'Klaida įkeliant sistemos nustatymus',
      type: 'alert-error'
    });

    loggingStore.error('Klaida įkeliant sistemos nustatymus', {
      component: 'SystemSettings',
      action: 'load_settings_failed',
      error: error.response?.data || error.message
    });
  }
};

/**
 * Opens the setting edit modal for the specified setting
 * @param {Object} setting - Setting to edit
 */
const editSetting = (setting) => {
  currentSetting.value = { ...setting };

  if (isBooleanSetting(setting.setting_key)) {
    editedSettingValue.value = setting.setting_value === 'true';
  } else {
    editedSettingValue.value = setting.setting_value;
  }

  editingSettings.value = true;
};

/**
 * Saves the edited setting value to the server
 */
const saveSetting = async () => {
  try {
    if (!currentSetting.value) return;

    const settingToSave = { ...currentSetting.value };
    if (typeof editedSettingValue.value === 'boolean') {
      settingToSave.setting_value = editedSettingValue.value ? 'true' : 'false';
    } else {
      settingToSave.setting_value = editedSettingValue.value;
    }

    await api.put('/system-settings', settingToSave);

    const index = settings.value.findIndex(s => s.id === settingToSave.id);
    if (index !== -1) {
      settings.value[index] = settingToSave;
    }

    slogStore.addToast({
      message: `Nustatymas "${getDisplayName(settingToSave.setting_key)}" sėkmingai atnaujintas`,
      type: 'alert-success'
    });

    loggingStore.info('Sistemos nustatymas atnaujintas', {
      component: 'SystemSettings',
      action: 'update_setting',
      key: settingToSave.setting_key,
      newValue: settingToSave.setting_value
    });

    editingSettings.value = false;
  } catch (error) {
    slogStore.addToast({
      message: 'Klaida atnaujinant nustatymą',
      type: 'alert-error'
    });

    loggingStore.error('Klaida atnaujinant nustatymą', {
      component: 'SystemSettings',
      action: 'update_setting_failed',
      key: currentSetting.value?.setting_key,
      error: error.response?.data || error.message
    });
  }
};

/**
 * Updates a boolean setting directly from toggle control
 * @param {Object} setting - Setting to update
 * @param {Event} event - Change event
 */
const updateBooleanSetting = async (setting, event) => {
  try {
    const isChecked = event.target.checked;
    const settingToUpdate = { ...setting, setting_value: isChecked ? 'true' : 'false' };

    await api.put('/system-settings', settingToUpdate);

    const index = settings.value.findIndex(s => s.id === setting.id);
    if (index !== -1) {
      settings.value[index] = settingToUpdate;
    }

    slogStore.addToast({
      message: `Nustatymas "${getDisplayName(setting.setting_key)}" sėkmingai atnaujintas`,
      type: 'alert-success'
    });

    loggingStore.info('Sistemos nustatymas atnaujintas', {
      component: 'SystemSettings',
      action: 'toggle_setting',
      key: setting.setting_key,
      newValue: settingToUpdate.setting_value
    });
  } catch (error) {
    // Revert UI change on error
    event.target.checked = !event.target.checked;

    slogStore.addToast({
      message: 'Klaida atnaujinant nustatymą',
      type: 'alert-error'
    });

    loggingStore.error('Klaida atnaujinant nustatymą', {
      component: 'SystemSettings',
      action: 'toggle_setting_failed',
      key: setting.setting_key,
      error: error.response?.data || error.message
    });
  }
};

/**
 * Loads blacklisted IP addresses from the server
 */
const loadBlacklistedIPs = async () => {
  try {
    const response = await api.get('/blacklisted-ips');
    blacklistedIPs.value = response.data;

    loggingStore.info('IP juodasis sąrašas užkrautas', {
      component: 'SystemSettings',
      action: 'load_blacklist',
      count: response.data.length
    });
  } catch (error) {
    slogStore.addToast({
      message: 'Klaida įkeliant IP juodąjį sąrašą',
      type: 'alert-error'
    });

    loggingStore.error('Klaida įkeliant IP juodąjį sąrašą', {
      component: 'SystemSettings',
      action: 'load_blacklist_failed',
      error: error.response?.data || error.message
    });
  }
};

/**
 * Adds a new IP address to the blacklist
 */
const addBlacklistedIP = async () => {
  try {
    if (!newIP.value.ip_address) {
      slogStore.addToast({
        message: 'IP adresas negali būti tuščias',
        type: 'alert-warning'
      });
      return;
    }

    const ipRegex = /^(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/;
    if (!ipRegex.test(newIP.value.ip_address)) {
      slogStore.addToast({
        message: 'Neteisingas IP adreso formatas',
        type: 'alert-warning'
      });
      return;
    }

    await api.post('/blacklisted-ips', newIP.value);
    await loadBlacklistedIPs();

    newIP.value = {
      ip_address: '',
      reason: ''
    };

    slogStore.addToast({
      message: 'IP adresas sėkmingai pridėtas į juodąjį sąrašą',
      type: 'alert-success'
    });

    loggingStore.info('IP adresas pridėtas į juodąjį sąrašą', {
      component: 'SystemSettings',
      action: 'add_blacklisted_ip',
      ip: newIP.value.ip_address
    });
  } catch (error) {
    slogStore.addToast({
      message: 'Klaida pridedant IP adresą į juodąjį sąrašą',
      type: 'alert-error'
    });

    loggingStore.error('Klaida pridedant IP adresą į juodąjį sąrašą', {
      component: 'SystemSettings',
      action: 'add_blacklisted_ip_failed',
      ip: newIP.value.ip_address,
      error: error.response?.data || error.message
    });
  }
};

/**
 * Opens confirmation dialog for IP deletion
 * @param {Object} ip - IP address to delete
 */
const confirmDeleteIP = (ip) => {
  ipToDelete.value = ip;
};

/**
 * Removes an IP address from the blacklist
 */
const removeBlacklistedIP = async () => {
  if (!ipToDelete.value) return;

  try {
    await api.delete('/blacklisted-ips', { data: { id: ipToDelete.value.id } });

    blacklistedIPs.value = blacklistedIPs.value.filter(ip => ip.id !== ipToDelete.value.id);

    slogStore.addToast({
      message: `IP adresas ${ipToDelete.value.ip_address} sėkmingai pašalintas iš juodojo sąrašo`,
      type: 'alert-success'
    });

    loggingStore.info('IP adresas pašalintas iš juodojo sąrašo', {
      component: 'SystemSettings',
      action: 'remove_blacklisted_ip',
      ip: ipToDelete.value.ip_address
    });

    ipToDelete.value = null;
  } catch (error) {
    slogStore.addToast({
      message: 'Klaida šalinant IP adresą iš juodojo sąrašo',
      type: 'alert-error'
    });

    loggingStore.error('Klaida šalinant IP adresą iš juodojo sąrašo', {
      component: 'SystemSettings',
      action: 'remove_blacklisted_ip_failed',
      ip: ipToDelete.value?.ip_address,
      error: error.response?.data || error.message
    });
  }
};

// Load data on component mount
onMounted(async () => {
  await Promise.all([
    loadSettings(),
    loadBlacklistedIPs()
  ]);

  loggingStore.info('Sistemos nustatymų puslapis užkrautas', {
    component: 'SystemSettings',
    timestamp: new Date().toISOString()
  });
});
</script>

<style scoped>
.card-body {
  padding: 1.5rem;
}

.modal-box {
  max-width: 32rem;
}
</style>
