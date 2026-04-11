import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { apiFetch } from '../services/logout/autoLogoutRedirect.ts'

const routes: RouteRecordRaw[] = [
    {
        path: '/',
        component: () => import('../views/userAuthentication.vue')
    },
    {
        path: '/print',
        component: () => import('../views/labelPrint.vue')
    },
    {
        path: '/dashboard',
        component: () => import('../views/mainDashboard.vue')
    }
]

const router = createRouter({
    history: createWebHistory(),
    routes
})

router.beforeEach(async (to, _from, next) => {
    if (to.path === '/' && !to.query.logout) {
        try {
            const res = await apiFetch('/api/auth/me', {})
            if (res && res.ok) {
                next('/dashboard')
                return
            }
        } catch {
            // no valid session
        }
    }
    next()
})

export default router