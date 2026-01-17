import { defineStore } from 'pinia'
import { ref } from 'vue'
import { domainApi, credentialApi } from '@/api'

export const useDomainStore = defineStore('domain', () => {
  const domains = ref([])
  const credentials = ref([])
  const loading = ref(false)
  const currentDomain = ref('')
  const domainBookmarks = ref([])

  async function fetchDomains() {
    loading.value = true
    try {
      const res = await domainApi.list()
      domains.value = res.domains || []
    } finally {
      loading.value = false
    }
  }

  async function fetchDomainBookmarks(domain) {
    currentDomain.value = domain
    const res = await domainApi.getBookmarks(domain)
    domainBookmarks.value = res.bookmarks || []
  }

  async function fetchCredentials(domain) {
    credentials.value = await credentialApi.getByDomain(domain) || []
  }

  async function createCredential(data) {
    const cred = await credentialApi.create(data)
    if (currentDomain.value === data.domain) {
      await fetchCredentials(data.domain)
    }
    return cred
  }

  async function updateCredential(id, data) {
    await credentialApi.update(id, data)
    if (currentDomain.value) {
      await fetchCredentials(currentDomain.value)
    }
  }

  async function deleteCredential(id) {
    await credentialApi.delete(id)
    if (currentDomain.value) {
      await fetchCredentials(currentDomain.value)
    }
  }

  return {
    domains,
    credentials,
    loading,
    currentDomain,
    domainBookmarks,
    fetchDomains,
    fetchDomainBookmarks,
    fetchCredentials,
    createCredential,
    updateCredential,
    deleteCredential
  }
})
