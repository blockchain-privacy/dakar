<!-- SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no> -->
<!-- SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no> -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div>
    <!-- single root so it can be transitioned -->
    <v-navigation-drawer
      v-model="drawerModel"
      :style="{'position':$vuetify.display.mobile?'fixed':'absolute', 'max-height':$vuetify.display.mobile?'300px': undefined}"
      :location="$vuetify.display.mobile?'bottom':undefined"
      :temporary="$vuetify.display.mobile"
    >
      <v-list-item :to="{name: ROUTE_NAME_WIKI_ROOT}">
        <template #prepend>
          <v-icon>{{ mdiMagnify }}</v-icon>
        </template>
        <v-list-item-title class="text-title-large">
          Search Wiki
        </v-list-item-title>
      </v-list-item>
      <v-divider />
      <v-list
        v-if="fileSet"
        nav
      >
        <div
          v-for="fileItem in fileHierarchy"
          :key="fileItem.name"
        >
          <v-list-group
            v-if="fileItem.items"
            v-model="fileItem.active"
          >
            <template #activator="{props}">
              <v-list-item
                v-bind="props"
                :title="fileItem.name"
                :prepend-icon="mdiFolder"
              />
            </template>
            <!-- manually set padding of child item, as vuetify default is too large -->
            <v-list-item
              v-for="child in fileItem.items"
              :key="child.title"
              :to="{name: ROUTE_NAME_WIKI, params: { file: child.path }}"
              :title="child.name"
              :prepend-icon="mdiFileDocument"
              style="padding-left: 30px !important"
            />
          </v-list-group>
          <v-list-item
            v-else
            :to="{name: ROUTE_NAME_WIKI, params: { file: fileItem.path }}"
            :title="fileItem.name"
          >
            <template #prepend>
              <v-icon>{{ mdiFileDocument }}</v-icon>
            </template>
          </v-list-item>
        </div>
      </v-list>
      <v-skeleton-loader
        v-else
        type="list-item-three-line,list-item-three-line,list-item-three-line"
      />
    </v-navigation-drawer>
    <div class="position-relative mt-5">
      <v-btn
        :icon="mdiMenu"
        variant="text"
        class="position-absolute"
        style="top: -20px"
        @click="drawerModel = !drawerModel"
      />
      <fade-transition>
        <v-card
          v-if="showRootPage"
          variant="text"
          max-width="700px"
          class="mx-auto"
        >
          <v-card-text>
            <v-text-field
              v-model="query"
              label="Search wiki pages"
              hide-details
              variant="solo"
              @update:model-value="queueSearch"
            />
            <alert :text="errorMsg" />
            <v-expand-transition>
              <template v-if="hasSearched">
                <div
                  v-if="searchResults.length ===0"
                  class="text-center text-body-large mt-3"
                >
                  No results
                </div>
                <div v-else>
                  <v-card
                    v-for="(item) in searchResults"
                    :key="item.path"
                    class="my-4"
                  >
                    <v-card-title class="d-flex align-center">
                      <v-icon :icon="mdiFileDocument" />
                      <router-link :to="{name: ROUTE_NAME_WIKI, params: {file: item.path}}">
                        {{ item.title }}
                      </router-link>
                    </v-card-title>
                    <v-card-text v-if="item.fragment">
                      <!-- html is loaded from safe source -->
                      <!-- eslint-disable vue/no-v-html -->
                      <div
                        class="text-title-small"
                        v-html="item.fragment"
                      />
                    </v-card-text>
                  </v-card>
                </div>
              </template>
            </v-expand-transition>
          </v-card-text>
        </v-card>
        <template v-else>
          <!-- html is loaded from safe source -->
          <!-- eslint-disable vue/no-v-html -->
          <div
            v-if="fileHTML"
            :class="{'wikiFileContentFullSize': $vuetify.display.mobile,
                     'wikiFileContent': !$vuetify.display.mobile, 'mx-auto': !$vuetify.display.smAndDown,'text-break':true}"
            v-html="fileHTML"
          />
          <v-skeleton-loader
            v-else
            type="article"
            :class="{'wikiFileContentFullSize': $vuetify.display.mobile,
                     'wikiFileContent': !$vuetify.display.mobile, 'mx-auto': !$vuetify.display.smAndDown}"
          />
        </template>
      </fade-transition>
    </div>
  </div>
</template>

<script setup>
import {
	mdiFileDocument,
	mdiFolder,
	mdiMagnify,
	mdiMenu,
} from '@mdi/js';
import {
	computed,
	inject,
	onMounted,
	onUnmounted,
	ref,
	watch,
} from 'vue';
import {useRoute} from 'vue-router';
import {useDisplay} from 'vuetify';
import FadeTransition from '../common/FadeTransition.vue';
import {PAGE_TITLE, ROUTE_NAME_WIKI, ROUTE_NAME_WIKI_ROOT} from '@/constants';
import Alert from '@/components/common/Alert.vue';

const route = useRoute();
const wikiapi = inject('wikiapi');
const display = useDisplay();

const errorMsg = ref('');
const drawerModel = ref(!display.mobile.value);
const fileHTML = ref('');

// FileSet is going to hold a set with all possible file paths
const fileSet = ref(null);

// IsRoot determines if the root page of the wiki is shown
const showRootPage = ref(true);
const query = ref(null);
const searchResults = ref([]);
// Set to true if search has been executed at least once
const hasSearched = ref(false);
let searchTimer = null;

// Computed

// fileHierarchy returns a file hierarchy based on the given directories.
// convert map to array of objects
// result:
// [
//   {
//     "name": "Index",
//     "items": null,
//     "path": "index.md"
//   },
//   {
//     "name": "Dash",
//     "items": [
//       {
//         "name": "Destination",
//         "path": "dash/destination.md"
//       },
//       {
//         "name": "Mixing",
//         "path": "dash/mixing.md"
//       },
//     ],
//     "path": "dash/destination.md"
//   }
// ]
const fileHierarchy = computed(() => {
	if (fileSet.value === null) {
		return [];
	}

	const hierarchy = new Map();

	fileSet.value.forEach(d => {
		const pathParts = d.split('/');

		if (pathParts.length > 2) {
			// Only a depth of 2 is supported
			return;
		}

		let [directory, fileName] = pathParts;

		directory = cleanName(directory);

		const itemProps = {items: null, path: d};

		if (fileName) {
			fileName = cleanName(fileName);
			let props = itemProps;

			if (hierarchy.has(directory)) {
				props = hierarchy.get(directory);
			}

			props.items ??= [];

			props.items.push({name: fileName, path: d});
			hierarchy.set(directory, props);
		} else {
			hierarchy.set(directory, itemProps);
		}
	});

	const hierarchyArray = [];
	hierarchy.forEach((props, directory) => {
		const item = {name: directory, ...props};
		hierarchyArray.push(item);
	});

	return hierarchyArray;
});

const filepathToFile = computed(() => {
	const fileMap = new Map();

	if (fileSet.value === null) {
		return fileMap;
	}

	fileSet.value.forEach(d => {
		const pathParts = d.split('/');

		if (pathParts.length > 2) {
			// Only a depth of 2 is supported
			return;
		}

		const [directory, fileName] = pathParts;

		if (fileName) {
			fileMap.set(d, {name: cleanName(fileName), directory: capitalize(directory)});
		} else {
			// First split is the actual file path
			fileMap.set(d, {name: cleanName(directory), directory: ''});
		}
	});

	return fileMap;
});

// SeparateWords adds a space before each capitalized letter and numbers
function separateWords(string) {
	return string.replaceAll(/([A-Z])/gv, ' $1').replaceAll(/(\d+(\.\d+)?)/gv, ' $1').trim();
}

// Capitalize capitalizes the given string
function capitalize(string) {
	return string[0].toUpperCase() + string.slice(1);
}

// CleanName capitalizes the given string and remove the '.md' postfix
function cleanName(fileName) {
	return capitalize(separateWords(fileName)).replace('.md', '').replace('-', ' ');
}

function setErrorMessage(msg) {
	errorMsg.value = msg;
}

async function getFileIndex() {
	errorMsg.value = '';
	try {
		const response = await wikiapi.indexGet();
		if (response.index) {
			fileSet.value = new Set(response.index);
		}
	} catch (error) {
		setErrorMessage(error);
	}
}

async function getFile(filePath) {
	// Only try to get file if it is in list of known files
	if (fileSet.value === null || !fileSet.value.has(filePath)) {
		return;
	}

	fileHTML.value = '';
	errorMsg.value = '';
	document.title = `Wiki - ${cleanName(filePath)} - ${PAGE_TITLE}`;

	try {
		const response = await wikiapi.fileFileNameGet({fileName: filePath});

		if (response.html) {
			fileHTML.value = response.html;
		}
	} catch (error) {
		setErrorMessage(error);
	}
}

function queueSearch(q) {
	if (searchTimer !== null) {
		clearTimeout(searchTimer);
	}

	searchTimer = setTimeout(search, 700, q);
}

async function search(q) {
	// Only search if files were loaded
	if (filepathToFile.value.size === 0) {
		return;
	}

	if (!q) {
		return;
	}

	q = q.trim();

	if (q.length < 3) {
		return;
	}

	hasSearched.value = true;
	let ret = [];
	errorMsg.value = '';
	try {
		const response = await wikiapi.searchPost({query: {query: q}});

		if (response.searchResults && response.searchResults.length > 0) {
			ret = response.searchResults
				.map(f => {
					const file = filepathToFile.value.get(f.filename);
					let title;
					if (file) {
						title = file.name;
						if (file.directory) {
							title = file.directory + ' - ' + title;
						}
					}

					return {title, path: f.filename, fragment: f.fragment};
				})
				.filter(d => Boolean(d.title));
		}
	} catch (error) {
		setErrorMessage(error);
	}

	searchResults.value = ret;
}

watch(route, () => {
	if (route.params.file) {
		showRootPage.value = false;
		getFile(route.params.file);
	} else {
		showRootPage.value = true;
		document.title = `Wiki - ${PAGE_TITLE}`;
	}
});

watch(display.mobile, newVal => {
	if (!newVal) {
		// Show drawer on desktop;
		drawerModel.value = true;
	}
});

// Hooks
onMounted(async () => {
	document.title = `Wiki - ${PAGE_TITLE}`;

	if (route.params.file) {
		showRootPage.value = false;
	}

	await getFileIndex();

	if (!showRootPage.value) {
		await getFile(route.params.file);
	}
});

onUnmounted(() => {
	if (searchTimer !== null) {
		clearTimeout(searchTimer);
	}
});

</script>

<style scoped>
.wikiFileContent :deep(img){
  display:block;
  max-width:60%;
  margin-left: auto;
  margin-right: auto;
}

.wikiFileContentFullSize :deep(img) {
  max-width: 100%;
}

:deep(img){
  margin-top: 30px;
  margin-bottom: 30px;
}

 :deep(li){
  margin-left: 15px;
}

 :deep(h2){
   margin-bottom: 15px;
 }

:deep(h3){
  margin-bottom: 10px;
}

/* <em> after <img> */
:deep(img ~ em){
  display:block;
  text-align: center;
}

.wikiFileContent{
  margin-bottom: 50px;
  max-width:900px
}

:deep(p) {
  margin-bottom: 10px;
}

:deep(ul) {
  margin-bottom: 10px;
  margin-left: 20px;
}

:deep(li) {
  margin-bottom: 5px;
}

.wikiFileContentFullSize{
  margin: 0 10px 60px 10px;
}
</style>
