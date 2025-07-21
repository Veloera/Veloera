/*
Copyright (c) 2025 Tethys Plex

This file is part of Veloera.

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.
*/
import React, { lazy, Suspense, useContext, useEffect } from 'react';
import { Route, Routes, useLocation } from 'react-router-dom';
import Loading from './components/Loading';
import User from './pages/User';
import { PrivateRoute } from './components/PrivateRoute';
import RegisterForm from './components/RegisterForm';
import LoginForm from './components/LoginForm';
import NotFound from './pages/NotFound';
import Setting from './pages/Setting';
import EditUser from './pages/User/EditUser';
import PasswordResetForm from './components/PasswordResetForm';
import PasswordResetConfirm from './components/PasswordResetConfirm';
import Channel from './pages/Channel';
import Token from './pages/Token';
import EditChannel from './pages/Channel/EditChannel';
import Redemption from './pages/Redemption';
import TopUp from './pages/TopUp';
import Log from './pages/Log';
import Chat from './pages/Chat';
import Chat2Link from './pages/Chat2Link';
import { Layout } from '@douyinfe/semi-ui';
import Midjourney from './pages/Midjourney';
import Pricing from './pages/Pricing/index.js';
import Task from './pages/Task/index.js';
import Playground from './pages/Playground/Playground.js';
import OAuth2Callback from './components/OAuth2Callback.js';
import PersonalSetting from './components/PersonalSetting.js';
import Setup from './pages/Setup/index.js';
import SetupCheck from './components/SetupCheck';
import Inbox from './pages/Inbox';
import Message from './pages/Message';
import { AdminRoute } from './components/AdminRoute';

const Home = lazy(() => import('./pages/Home'));
const Detail = lazy(() => import('./pages/Detail'));
const About = lazy(() => import('./pages/About'));

function App() {
  const location = useLocation();

  return (
    <SetupCheck>
      <Routes>
        <Route
          path='/'
          element={
            <Suspense fallback={<Loading></Loading>} key={location.pathname}>
              <Home />
            </Suspense>
          }
        />
        <Route
          path='/setup'
          element={
            <Suspense fallback={<Loading></Loading>} key={location.pathname}>
              <Setup />
            </Suspense>
          }
        />
        <Route
          path='/admin/channels'
          element={
            <PrivateRoute>
              <Channel />
            </PrivateRoute>
          }
        />
        <Route
          path='/channel/edit/:id'
          element={
            <Suspense fallback={<Loading></Loading>} key={location.pathname}>
              <EditChannel />
            </Suspense>
          }
        />
        <Route
          path='/channel/add'
          element={
            <Suspense fallback={<Loading></Loading>} key={location.pathname}>
              <EditChannel />
            </Suspense>
          }
        />
        <Route
          path='/app/tokens'
          element={
            <PrivateRoute>
              <Token />
            </PrivateRoute>
          }
        />
        <Route
          path='/playground'
          element={
            <PrivateRoute>
              <Playground />
            </PrivateRoute>
          }
        />
        <Route
          path='/admin/coupons'
          element={
            <PrivateRoute>
              <Redemption />
            </PrivateRoute>
          }
        />
        <Route
          path='/admin/users'
          element={
            <PrivateRoute>
              <User />
            </PrivateRoute>
          }
        />
        <Route
          path='/user/edit/:id'
          element={
            <Suspense fallback={<Loading></Loading>} key={location.pathname}>
              <EditUser />
            </Suspense>
          }
        />
        <Route
          path='/user/edit'
          element={
            <Suspense fallback={<Loading></Loading>} key={location.pathname}>
              <EditUser />
            </Suspense>
          }
        />
        <Route
          path='/user/reset'
          element={
            <Suspense fallback={<Loading></Loading>} key={location.pathname}>
              <PasswordResetConfirm />
            </Suspense>
          }
        />
        <Route
          path='/login'
          element={
            <Suspense fallback={<Loading></Loading>} key={location.pathname}>
              <LoginForm />
            </Suspense>
          }
        />
        <Route
          path='/register'
          element={
            <Suspense fallback={<Loading></Loading>} key={location.pathname}>
              <RegisterForm />
            </Suspense>
          }
        />
        <Route
          path='/reset'
          element={
            <Suspense fallback={<Loading></Loading>} key={location.pathname}>
              <PasswordResetForm />
            </Suspense>
          }
        />
        <Route
          path='/oauth/github'
          element={
            <Suspense fallback={<Loading></Loading>} key={location.pathname}>
              <OAuth2Callback type='github'></OAuth2Callback>
            </Suspense>
          }
        />
        <Route
          path='/oauth/oidc'
          element={
            <Suspense fallback={<Loading></Loading>}>
              <OAuth2Callback type='oidc'></OAuth2Callback>
            </Suspense>
          }
        />
        <Route
          path='/oauth/linuxdo'
          element={
            <Suspense fallback={<Loading></Loading>} key={location.pathname}>
              <OAuth2Callback type='linuxdo'></OAuth2Callback>
            </Suspense>
          }
        />
        <Route
          path='/admin/settings'
          element={
            <PrivateRoute>
              <Suspense fallback={<Loading></Loading>} key={location.pathname}>
                <Setting />
              </Suspense>
            </PrivateRoute>
          }
        />
        <Route
          path='/app/me'
          element={
            <PrivateRoute>
              <Suspense fallback={<Loading></Loading>} key={location.pathname}>
                <PersonalSetting />
              </Suspense>
            </PrivateRoute>
          }
        />
        <Route
          path='/app/wallet'
          element={
            <PrivateRoute>
              <Suspense fallback={<Loading></Loading>} key={location.pathname}>
                <TopUp />
              </Suspense>
            </PrivateRoute>
          }
        />
        <Route
          path='/app/logs/api-usage'
          element={
            <PrivateRoute>
              <Log />
            </PrivateRoute>
          }
        />
        <Route
          path='/app/dashboard'
          element={
            <PrivateRoute>
              <Suspense fallback={<Loading></Loading>} key={location.pathname}>
                <Detail />
              </Suspense>
            </PrivateRoute>
          }
        />
        <Route
          path='/app/logs/drawing'
          element={
            <PrivateRoute>
              <Suspense fallback={<Loading></Loading>} key={location.pathname}>
                <Midjourney />
              </Suspense>
            </PrivateRoute>
          }
        />
        <Route
          path='/app/logs/tasks'
          element={
            <PrivateRoute>
              <Suspense fallback={<Loading></Loading>} key={location.pathname}>
                <Task />
              </Suspense>
            </PrivateRoute>
          }
        />
        <Route
          path='/pricing'
          element={
            <Suspense fallback={<Loading></Loading>} key={location.pathname}>
              <Pricing />
            </Suspense>
          }
        />
        <Route
          path='/about'
          element={
            <Suspense fallback={<Loading></Loading>} key={location.pathname}>
              <About />
            </Suspense>
          }
        />
        <Route
          path='/chat/:id?'
          element={
            <Suspense fallback={<Loading></Loading>} key={location.pathname}>
              <Chat />
            </Suspense>
          }
        />
        {/* 方便使用chat2link直接跳转聊天... */}
        <Route
          path='/chat2link'
          element={
            <PrivateRoute>
              <Suspense fallback={<Loading></Loading>} key={location.pathname}>
                <Chat2Link />
              </Suspense>
            </PrivateRoute>
          }
        />
        {/* User inbox routes */}
        <Route
          path='/app/inbox'
          element={
            <PrivateRoute>
              <Suspense fallback={<Loading></Loading>} key={location.pathname}>
                <Inbox />
              </Suspense>
            </PrivateRoute>
          }
        />
        <Route
          path='/app/inbox/:id'
          element={
            <PrivateRoute>
              <Suspense fallback={<Loading></Loading>} key={location.pathname}>
                <Inbox />
              </Suspense>
            </PrivateRoute>
          }
        />
        {/* Admin message management routes */}
        <Route
          path='/admin/messages'
          element={
            <AdminRoute>
              <Suspense fallback={<Loading></Loading>} key={location.pathname}>
                <Message />
              </Suspense>
            </AdminRoute>
          }
        />
        <Route path='*' element={<NotFound />} />
      </Routes>
    </SetupCheck>
  );
}

export default App;
