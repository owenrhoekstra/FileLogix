import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { apiFetch } from '../services/fetch/statusCodeChecks.ts'

const routes: RouteRecordRaw[] = [
    {
        path: '/',
        component: () => import('../views/auth/userAuthentication.vue')
    },
    {
        path: '/dashboard',
        component: () => import('../views/viewRecords/mainRecordView.vue')
    },
    {
        path: '/old-dashboard',
        component: () => import('../views/mainDashboard.vue')
    },
    {
        path: '/setup',
        component: () => import('../views/auth/userSetup.vue')
    },
    {
        path: '/new-record',
        component: () => import('../views/addRecord/addNewRecord.vue')
    },
    {
        path: '/support',
        component: () => import('../views/support/appSupport.vue')
    },
    {
        path: '/settings',
        component: () => import('../views/optionsMenu/settingsView.vue')
    },
    {
        path: '/profile',
        component: () => import('../views/optionsMenu/profileView.vue')
    },
    {
        //path: '/records',
        //component: () => import('../views/viewRecords/mainRecordView.vue')
    },
    {
        path: '/print/:id',
        component: () => import('../views/addRecord/printView.vue')
    },
    {
        path: '/code-scan',
        component: () => import('../views/addRecord/filingView.vue')
    },
    { path: '/:pathMatch(.*)*', name: 'NotFound',
      component: () => import('../views/notFound.vue')
    }
]

const router = createRouter({
    history: createWebHistory(),
    routes
})

router.beforeEach(async (to, _from, next) => {
    // Public route — always allow
    if (to.path === '/') {
        // Redirect to dashboard or setup if already authenticated
        if (!to.query.logout) {
            try {
                const res = await apiFetch('/api/auth/me', {})
                if (res?.ok) {
                    const data = await res.json()
                    next(data.metadataComplete ? '/dashboard' : '/setup')
                    return
                }
            } catch { /* no session */ }
        }
        next()
        return
    }

    // All other routes require a valid session
    try {
        const res = await apiFetch('/api/auth/me', {})
        if (!res?.ok) {
            next('/')
            return
        }
        const data = await res.json()

        // Enforce setup before dashboard access
        if (to.path !== '/setup' && !data.metadataComplete) {
            next('/setup')
            return
        }

        // Prevent re-doing setup once complete
        if (to.path === '/setup' && data.metadataComplete) {
            next('/dashboard')
            return
        }

        next()
    } catch {
        next('/')
    }
})

export default router