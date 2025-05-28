import React, { lazy, Suspense, useContext, useEffect } from 'react';
import { Route, Routes, useLocation } from 'react-router-dom';
import Loading from './components/Loading';
import { PrivateRoute } from './components/PrivateRoute';
import SetupCheck from './components/SetupCheck';
import { Layout } from '@douyinfe/semi-ui';

const User = lazy(() => import('./pages/User'));
const RegisterForm = lazy(() => import('./components/RegisterForm'));
const LoginForm = lazy(() => import('./components/LoginForm'));
const NotFound = lazy(() => import('./pages/NotFound'));
const Setting = lazy(() => import('./pages/Setting'));
const EditUser = lazy(() => import('./pages/User/EditUser'));
const PasswordResetForm = lazy(() => import('./components/PasswordResetForm'));
const PasswordResetConfirm = lazy(
  () => import('./components/PasswordResetConfirm'),
);
const Channel = lazy(() => import('./pages/Channel'));
const Token = lazy(() => import('./pages/Token'));
const EditChannel = lazy(() => import('./pages/Channel/EditChannel'));
const Redemption = lazy(() => import('./pages/Redemption'));
const TopUp = lazy(() => import('./pages/TopUp'));
const Log = lazy(() => import('./pages/Log'));
const Chat = lazy(() => import('./pages/Chat'));
const Chat2Link = lazy(() => import('./pages/Chat2Link'));
const Midjourney = lazy(() => import('./pages/Midjourney'));
const Pricing = lazy(() => import('./pages/Pricing/index.js'));
const Task = lazy(() => import('./pages/Task/index.js'));
const Playground = lazy(() => import('./pages/Playground/Playground.js'));
const OAuth2Callback = lazy(() => import('./components/OAuth2Callback.js'));
const PersonalSetting = lazy(() => import('./components/PersonalSetting.js'));
const Setup = lazy(() => import('./pages/Setup/index.js'));

const Home = lazy(() => import('./pages/Home'));
const Detail = lazy(() => import('./pages/Detail'));
const About = lazy(() => import('./pages/About'));

function App() {
  const location = useLocation();

  return (
    <SetupCheck>
      <Suspense fallback={<Loading />} key={location.pathname}>
        <Routes>
          <Route path='/' element={<Home />} />
          <Route path='/setup' element={<Setup />} />
          <Route
            path='/channel'
            element={
              <PrivateRoute>
                <Channel />
              </PrivateRoute>
            }
          />
          <Route path='/channel/edit/:id' element={<EditChannel />} />
          <Route path='/channel/add' element={<EditChannel />} />
          <Route
            path='/token'
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
            path='/redemption'
            element={
              <PrivateRoute>
                <Redemption />
              </PrivateRoute>
            }
          />
          <Route
            path='/user'
            element={
              <PrivateRoute>
                <User />
              </PrivateRoute>
            }
          />
          <Route path='/user/edit/:id' element={<EditUser />} />
          <Route path='/user/edit' element={<EditUser />} />
          <Route path='/user/reset' element={<PasswordResetConfirm />} />
          <Route path='/login' element={<LoginForm />} />
          <Route path='/register' element={<RegisterForm />} />
          <Route path='/reset' element={<PasswordResetForm />} />
          <Route
            path='/oauth/github'
            element={<OAuth2Callback type='github' />}
          />
          <Route path='/oauth/oidc' element={<OAuth2Callback type='oidc' />} />
          <Route
            path='/oauth/linuxdo'
            element={<OAuth2Callback type='linuxdo' />}
          />
          <Route
            path='/setting'
            element={
              <PrivateRoute>
                <Setting />
              </PrivateRoute>
            }
          />
          <Route
            path='/personal'
            element={
              <PrivateRoute>
                <PersonalSetting />
              </PrivateRoute>
            }
          />
          <Route
            path='/topup'
            element={
              <PrivateRoute>
                <TopUp />
              </PrivateRoute>
            }
          />
          <Route
            path='/log'
            element={
              <PrivateRoute>
                <Log />
              </PrivateRoute>
            }
          />
          <Route
            path='/detail'
            element={
              <PrivateRoute>
                <Detail />
              </PrivateRoute>
            }
          />
          <Route
            path='/midjourney'
            element={
              <PrivateRoute>
                <Midjourney />
              </PrivateRoute>
            }
          />
          <Route
            path='/task'
            element={
              <PrivateRoute>
                <Task />
              </PrivateRoute>
            }
          />
          <Route path='/pricing' element={<Pricing />} />
          <Route path='/about' element={<About />} />
          <Route path='/chat/:id?' element={<Chat />} />
          {/* 方便使用chat2link直接跳转聊天... */}
          <Route
            path='/chat2link'
            element={
              <PrivateRoute>
                <Chat2Link />
              </PrivateRoute>
            }
          />
          <Route path='*' element={<NotFound />} />
        </Routes>
      </Suspense>
    </SetupCheck>
  );
}

export default App;
