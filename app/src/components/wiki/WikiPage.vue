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
          <v-list-item :to="{name: routeWikiRoot}">
            <template #prepend>
              <v-icon>{{ icons.mdiBookOpen }}</v-icon>
            </template>
            <v-list-item-title class="text-h6">
              Wiki
            </v-list-item-title>
          </v-list-item>
          <v-divider />
          <v-list
            v-if="fileSet"
            density="compact"
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
                    :prepend-icon="icons.mdiBook"
                  />
                </template>
                <v-list-item
                  v-for="child in fileItem.items"
                  :key="child.title"
                  :to="{name: routeWiki, params: { file: child.path }}"
                  :title="child.name"
                >
                  <template #prepend>
                    <v-icon>{{ icons.mdiBook }}</v-icon>
                  </template>
                </v-list-item>
              </v-list-group>
              <v-list-item
                v-else
                :to="{name: routeWiki, params: { file: fileItem.path }}"
                :title="fileItem.name"
              >
                <template #prepend>
                  <v-icon>{{ icons.mdiBook }}</v-icon>
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
                :append-icon="icons.mdiMagnify"
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

<script>
import {mdiBookOpen, mdiBook, mdiMagnify} from '@mdi/js';
import {
	PAGE_TITLE, ROUTE_NAME_WIKI, ROUTE_NAME_WIKI_ROOT,
} from '@/constants';
import FadeTransition from '../common/FadeTransition.vue';

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
function getFileHierarchy(fileSet) {
	if (fileSet === null) {
		return [];
	}

	const hierarchy = new Map();

	fileSet.forEach(d => {
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

			if (!props.items) {
				props.items = [];
			}

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

export default {
	name: 'WikiPage',
	components: {FadeTransition},
	data() {
		return {
			icons: {
				mdiBookOpen, mdiBook, mdiMagnify,
			},
			routeWiki: ROUTE_NAME_WIKI,
			routeWikiRoot: ROUTE_NAME_WIKI_ROOT,
			fileHTML: '',
			// FileSet will hold a set with all possible file paths
			fileSet: null,
			// IsRoot determines of the root page of the wiki is shown
			showRootPage: true,
			query: null,
		};
	},
	computed: {
		fileHierarchy() {
			return getFileHierarchy(this.fileSet);
		},
		// NamePathPairs returns an array of name and
		// path pairs [{name: filename, path: filename.md}, ...]
		namePathPairs() {
			const pairs = [];

			this.fileHierarchy.forEach(d => {
				if (d.items) {
					d.items.forEach(l => {
						pairs.push(l);
					});
				} else {
					pairs.push(d);
				}
			});

			return pairs;
		},
	},
	watch: {
		$route() {
			if (this.$route.params.file) {
				this.showRootPage = false;
				this.getFile(this.$route.params.file);
			} else {
				this.showRootPage = true;
				document.title = `Wiki - ${PAGE_TITLE}`;
			}
		},
	},
	async mounted() {
		document.title = `Wiki - ${PAGE_TITLE}`;

		if (this.$route.params.file) {
			this.showRootPage = false;
		}

		await this.getFileIndex();

		if (!this.showRootPage) {
			await this.getFile(this.$route.params.file);
		}
	},
	methods: {
		setErrorMessage(msg) {
			this.$store.dispatch('addMessage', {text: msg, type: 'error', temporary: true, category: this.$route.name});
		},
		async getFileIndex() {
			try {
				const response = await this.wikiapi.indexGet();
				if (response.index) {
					this.fileSet = new Set(response.index);
				}
			} catch (e) {
				this.setErrorMessage(e);
			}
		},
		async getFile(filePath) {
			// Only try to get file if it is in list of known files
			if (this.fileSet === null || !this.fileSet.has(filePath)) {
				return;
			}

			this.fileHTML = '';

			document.title = `Wiki - ${cleanName(filePath)} - ${PAGE_TITLE}`;

			try {
				const response = await this.wikiapi.fileFileNameGet({fileName: filePath});

				if (response.html) {
					this.fileHTML = response.html;
				}
			} catch (e) {
				this.setErrorMessage(e);
			}
		},
		navigateToWikiPage() {
			if (!this.query) {
				return;
			}

			this.$router.push({name: this.routeWiki, params: {file: this.query}});
		},
	},
};
</script>

<style scoped>

.wikiFileContent :deep( img ) {
  max-width: 40%;
}

.wikiFileContentFullSize :deep( img ) {
  max-width: 100%;
}

</style>
