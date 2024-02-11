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
            :nav="true"
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
            :flat="true"
          >
            <v-card-text>
              <v-autocomplete
                v-model="query"
                :items="namePathPairs"
                item-title="name"
                item-value="path"
                label="Search for wiki pages"
                variant="outlined"
                :append-icon="mdiMagnify"
                @update:model-value="navigateToWikiPage"
                @keydown.enter="navigateToWikiPage"
                @click:append="navigateToWikiPage"
              />
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
import {mdiBook, mdiBookOpen, mdiMagnify} from '@mdi/js';
import {PAGE_TITLE, ROUTE_NAME_WIKI, ROUTE_NAME_WIKI_ROOT} from '@/constants';
import FadeTransition from '../common/FadeTransition.vue';
import {
	computed, inject, onMounted, ref, watch,
} from 'vue';
import {useRoute, useRouter} from 'vue-router';
import {useMsgStore} from '@/pinia/msg';

const route = useRoute();
const router = useRouter();
const wikiapi = inject('wikiapi');
const msgStore = useMsgStore();

const fileHTML = ref('');

// FileSet is going to hold a set with all possible file paths
const fileSet = ref(null);

// IsRoot determines if the root page of the wiki is shown
const showRootPage = ref(true);
const query = ref(null);

// Computed
const fileHierarchy = computed(() => getFileHierarchy());

// NamePathPairs returns an array of name and
// path pairs [{name: filename, path: filename.md}, ...]
const namePathPairs = computed(() => {
	const pairs = [];

	fileHierarchy.value.forEach(d => {
		if (d.items) {
			d.items.forEach(l => {
				pairs.push(l);
			});
		} else {
			pairs.push(d);
		}
	});

	return pairs;
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

// GetFileHierarchy returns a file hierarchy based on the given directories.
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
function getFileHierarchy() {
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

function navigateToWikiPage() {
	if (!query.value) {
		return;
	}

	router.push({name: ROUTE_NAME_WIKI, params: {file: query.value}});
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

</script>

<style scoped>

.wikiFileContent :deep( img ) {
  max-width: 40%;
}

.wikiFileContentFullSize :deep( img ) {
  max-width: 100%;
}

</style>
