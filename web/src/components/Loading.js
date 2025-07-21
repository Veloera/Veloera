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
import React from 'react';

const Loading = ({ prompt: name = 'page' }) => {
  return (
    <div style={{
      display: 'flex',
      justifyContent: 'center',
      alignItems: 'center',
      minHeight: '100vh'
    }}>
      <div style={{
        width: '200px',
        height: '200px',
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center'
      }}>
        <svg viewBox="0 0 140 140" fill="none" style={{
          width: '140px',
          height: '140px'
        }}>
          <title>Loading {name}</title>
          <defs>
            <linearGradient id="gradient" x1="70" y1="0" x2="70" y2="140" gradientUnits="userSpaceOnUse">
              <stop offset="0%" stopColor="var(--primary, #AAE4FF)" />
              <stop offset="100%" stopColor="var(--primary-foreground, #38BDF8)" />
            </linearGradient>
          </defs>
          <g style={{
            animation: 'rotate 6s cubic-bezier(0.68, -0.55, 0.27, 1.55) infinite',
            transformOrigin: 'center'
          }}>
            <g style={{
              animation: 'pulsateLeft 6s ease-in-out infinite',
              transformOrigin: 'center'
            }}>
              <path 
                d="M70,70 m0,-52 a52,52 0 1,1 0,104 a52,52 0 1,1 0,-104 z M70,70 m0,-32 a32,32 0 0,0 0,64 a32,32 0 0,0 0,-64 z" 
                fill="url(#gradient)" 
                fillRule="evenodd"
                clipRule="evenodd"
              />
            </g>
          </g>
          <g style={{
            animation: 'rotate 6s cubic-bezier(0.68, -0.55, 0.27, 1.55) infinite',
            transformOrigin: 'center'
          }}>
            <g style={{
              animation: 'pulsateRight 6s ease-in-out infinite',
              transformOrigin: 'center'
            }}>
              <path 
                d="M70,70 m-52,0 a52,52 0 1,0 104,0 a52,52 0 1,0 -104,0 z M70,70 m-32,0 a32,32 0 1,1 64,0 a32,32 0 1,1 -64,0 z" 
                fill="url(#gradient)" 
                fillRule="evenodd"
                clipRule="evenodd"
              />
            </g>
          </g>
        </svg>
        <style>
          {`
            @keyframes rotate {
              0%, 70% { transform: rotate(0deg); }
              30%, 40% { transform: rotate(720deg); }
            }
            
            @keyframes pulsateLeft {
              0%, 40%, 60%, 80%, 100% { transform: scale(1); }
              50% { transform: scale(1.15); }
              70% { transform: scale(1.1); }
              90% { transform: scale(1.05); }
            }
            
            @keyframes pulsateRight {
              0%, 50%, 70%, 90%, 100% { transform: scale(1); }
              40% { transform: scale(1.15); }
              60% { transform: scale(1.1); }
              80% { transform: scale(1.05); }
            }
          `}
        </style>
      </div>
    </div>
  );
};

export default Loading;
