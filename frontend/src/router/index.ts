import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
    {
        path: '/',
        component: () =>
            import('../views/userAuthentication.vue'),
    },
    {
        path: '/print',
        component: () =>
            import('../views/labelPrint.vue')
    },
    {
        path: '/dashboard',
        component: () =>
            import('../views/mainDashboard.vue')
    }
]

const router = createRouter({
    history: createWebHistory(),
    routes
})

export default router