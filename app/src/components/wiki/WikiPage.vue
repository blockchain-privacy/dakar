<template>
  <div
    class="fill-height"
    style="padding: 12px 10px 0 10px"
  >
    <v-row class="fill-height">
      <v-col
        v-if="!$vuetify.display.smAndDown"
        cols="auto"
        class="pa-0"
      >
        <v-navigation-drawer style="position:absolute">
          <v-list-item :to="{name: ROUTE_NAME_WIKI_ROOT}">
            <template #prepend>
              <v-icon>{{ mdiBookOpen }}</v-icon>
            </template>
            <v-list-item-title class="text-h6">
              Wiki
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
                    :prepend-icon="mdiBook"
                  />
                </template>
                <v-list-item
                  v-for="child in fileItem.items"
                  :key="child.title"
                  :to="{name: ROUTE_NAME_WIKI, params: { file: child.path }}"
                  :title="child.name"
                />
              </v-list-group>
              <v-list-item
                v-else
                :to="{name: ROUTE_NAME_WIKI, params: { file: fileItem.path }}"
                :title="fileItem.name"
              >
                <template #prepend>
                  <v-icon>{{ mdiBook }}</v-icon>
                </template>
              </v-list-item>
            </div>
          </v-list>
          <v-skeleton-loader
            v-else
            type="list-item-three-line,list-item-three-line,list-item-three-line"
          />
        </v-navigation-drawer>
      </v-col>
      <v-col class="fill-height mx-lg-16">
        <fade-transition>
          <v-card
            v-if="showRootPage"
            flat
          >
            <v-card-text>
              <v-text-field
                v-model="query"
                label="Search wiki pages"
                hide-details
                @update:model-value="queueSearch"
              />
              <v-expand-transition>
                <div
                  v-if="isSearching"
                  class="d-flex justify-center mt-3"
                >
                  <v-progress-circular indeterminate />
                </div>
                <template v-else-if="hasSearched">
                  <div
                    v-if="searchResults.length ===0"
                    class="text-center text-subtitle-1 mt-3"
                  >
                    No results
                  </div>
                  <v-list v-else>
                    <v-list-item
                      v-for="(item) in searchResults"
                      :key="item.path"
                      :prepend-icon="mdiBook"
                      :to="{name: ROUTE_NAME_WIKI, params: {file: item.path}}"
                    >
                      <v-list-item-title>{{ item.title }}</v-list-item-title>
                    </v-list-item>
                  </v-list>
                </template>
              </v-expand-transition>
            </v-card-text>
          </v-card>
          <template v-else>
            <!-- html is loaded from safe source -->
            <!-- eslint-disable vue/no-v-html -->
            <div
              v-if="fileHTML"
              :class="{'wikiFileContentFullSize': $vuetify.display.smAndDown,
                       'wikiFileContent': !$vuetify.display.smAndDown}"
              v-html="fileHTML"
            />
            <v-skeleton-loader
              v-else
              type="article"
            />
          </template>
        </fade-transition>
      </v-col>
    </v-row>
  </div>
</template>

<script setup>
import {mdiBook, mdiBookOpen} from '@mdi/js';
import {PAGE_TITLE, ROUTE_NAME_WIKI, ROUTE_NAME_WIKI_ROOT} from '@/constants';
import FadeTransition from '../common/FadeTransition.vue';
import {
	computed, inject, onMounted, onUnmounted, ref, watch,
} from 'vue';
import {useRoute} from 'vue-router';
import {useMsgStore} from '@/pinia/msg';

const route = useRoute();
const wikiapi = inject('wikiapi');
const msgStore = useMsgStore();

const fileHTML = ref('');

// FileSet is going to hold a set with all possible file paths
const fileSet = ref(null);

// IsRoot determines if the root page of the wiki is shown
const showRootPage = ref(true);
const query = ref(null);
const searchResults = ref([]);
// Set to true if search has been executed at least once
const hasSearched = ref(false);
const isSearching = ref(false);
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
//     "name": "TransactionTypes",
//     "items": [
//       {
//         "name": "Destination",
//         "path": "transactionTypes/destination.md"
//       },
//       {
//         "name": "Mixing",
//         "path": "transactionTypes/mixing.md"
//       },
//     ],
//     "path": "transactionTypes/destination.md"
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

const filepathToFilename = computed(() => {
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
			fileMap.set(d, cleanName(fileName));
		} else {
			// First split is the actual file path
			fileMap.set(d, cleanName(directory));
		}
	});

	return fileMap;
});

// SeparateWords adds a space before each capitalized letter
function separateWords(string) {
	return string.replace(/([A-Z])/g, ' $1').trim();
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
	msgStore.addMessage({
		text: msg, type: 'error', temporary: true, category: route.name,
	});
}

async function getFileIndex() {
	try {
		const response = await wikiapi.indexGet();
		if (response.index) {
			fileSet.value = new Set(response.index);
		}
	} catch (e) {
		setErrorMessage(e);
	}
}

async function getFile(filePath) {
	// Only try to get file if it is in list of known files
	if (fileSet.value === null || !fileSet.value.has(filePath)) {
		return;
	}

	fileHTML.value = '';

	document.title = `Wiki - ${cleanName(filePath)} - ${PAGE_TITLE}`;

	try {
		const response = await wikiapi.fileFileNameGet({fileName: filePath});

		if (response.html) {
			fileHTML.value = response.html;
		}
	} catch (e) {
		setErrorMessage(e);
	}
}

function queueSearch(q) {
	if (searchTimer !== null) {
		clearTimeout(searchTimer);
	}

	searchTimer = setTimeout(search, 700, q);
}

async function search(query) {
	// Only search if files where loaded
	if (filepathToFilename.value.size === 0) {
		return;
	}

	if (!query) {
		return;
	}

	query = query.trim();

	if (query.length < 3) {
		return;
	}

	isSearching.value = true;
	hasSearched.value = true;
	let ret = [];

	try {
		const response = await wikiapi.searchPost({query: {query}});

		if (response.files && response.files.length > 0) {
			ret = response.files.map(f => ({title: filepathToFilename.value.get(f), path: f})).filter(d => Boolean(d.title));
		}
	} catch (e) {
		setErrorMessage(e);
	}

	searchResults.value = ret;
	isSearching.value = false;
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

.wikiFileContent :deep( img ) {
  max-width: 40%;
}

.wikiFileContentFullSize :deep( img ) {
  max-width: 100%;
}

</style>
